// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful server shutdown.
	DefaultShutdownTimeout = 5 * time.Second
)

// Dependencies holds the HTTP server's dependencies
type Dependencies struct {
	fx.In

	Lifecycle    fx.Lifecycle
	Logger       logger.Interface
	Config       config.Interface
	Storage      storagetypes.Interface
	IndexManager api.IndexManager
	Context      context.Context
	Sources      sources.Interface
}

// Params holds the parameters required for running the HTTP server.
type Params struct {
	fx.In
	Server  *http.Server
	Logger  logger.Interface
	Storage storagetypes.Interface
	Config  config.Interface
}

// Server implements the HTTP server
type Server struct {
	config config.Interface
	Logger logger.Interface
	server *http.Server
}

// NewServer creates a new HTTP server instance
func NewServer(params Params) *Server {
	return &Server{
		config: params.Config,
		Logger: params.Logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.Logger.Info("Starting HTTP server", "addr", s.config.GetServerConfig().Address)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.Logger.Info("Stopping HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

// WaitForHealth waits for the server to become healthy
func WaitForHealth(ctx context.Context, serverAddr string, interval, timeout time.Duration) error {
	healthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	client := &http.Client{Timeout: interval}
	for {
		select {
		case <-healthCtx.Done():
			return fmt.Errorf("server failed to become healthy within %v", timeout)
		case <-ticker.C:
			url := fmt.Sprintf("http://%s/health", serverAddr)
			req, err := http.NewRequestWithContext(healthCtx, http.MethodGet, url, http.NoBody)
			if err != nil {
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

// Shutdown gracefully shuts down the server.
func Shutdown(ctx context.Context, server *http.Server) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("error shutting down server: %w", err)
	}
	return nil
}

// Cmd represents the HTTP server command
var Cmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Get logger from context
		loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
		log, ok := loggerValue.(logger.Interface)
		if !ok {
			return errors.New("logger not found in context or invalid type")
		}

		// Create Fx app with the module
		fxApp := fx.New(
			Module,
			fx.Provide(
				func() logger.Interface { return log },
			),
			fx.WithLogger(func() fxevent.Logger {
				return logger.NewFxLogger(log)
			}),
		)

		// Start the application
		log.Info("Starting HTTP server")
		startErr := fxApp.Start(cmd.Context())
		if startErr != nil {
			log.Error("Failed to start application", "error", startErr)
			return fmt.Errorf("failed to start application: %w", startErr)
		}

		// Wait for interrupt signal
		log.Info("Waiting for interrupt signal")
		<-cmd.Context().Done()

		// Stop the application
		log.Info("Stopping application")
		stopErr := fxApp.Stop(cmd.Context())
		if stopErr != nil {
			log.Error("Failed to stop application", "error", stopErr)
			return fmt.Errorf("failed to stop application: %w", stopErr)
		}

		log.Info("Application stopped successfully")
		return nil
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}

// Run starts the HTTP server.
func Run() error {
	// Create a new Fx application
	app := fx.New(
		Module,
	)

	// Create a context that will be cancelled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the application
	if err := app.Start(ctx); err != nil {
		return err
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop the application
	return app.Stop(ctx)
}
