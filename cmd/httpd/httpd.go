// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	signalHandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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

// serverState tracks the HTTP server's state
type serverState struct {
	mu       sync.Mutex
	started  bool
	shutdown bool
	// serverDone is closed when the server goroutine exits
	serverDone chan struct{}
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
			req, err := http.NewRequestWithContext(healthCtx, http.MethodGet, fmt.Sprintf("http://%s/health", serverAddr), nil)
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
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling
		handler := signalHandler.NewSignalHandler(logger.NewNoOp())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Track server state
		state := &serverState{
			serverDone: make(chan struct{}),
		}

		// Initialize the Fx application
		fxApp := fx.New(
			Module,
			fx.Provide(
				func() context.Context { return ctx },
			),
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
				// Update the signal handler with the real logger
				handler.SetLogger(p.Logger)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test storage connection
						if err := p.Storage.TestConnection(ctx); err != nil {
							return fmt.Errorf("failed to connect to storage: %w", err)
						}

						// Start HTTP server in background
						p.Logger.Info("Starting HTTP server...", "address", p.Server.Addr)
						state.mu.Lock()
						state.started = true
						state.mu.Unlock()

						// Create error channel to propagate server errors
						errChan := make(chan error, 1)
						go func() {
							defer close(state.serverDone)
							if err := p.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
								p.Logger.Error("HTTP server failed", "error", err)
								errChan <- fmt.Errorf("HTTP server failed: %w", err)
							}
						}()

						// Wait for server to be healthy
						if err := WaitForHealth(ctx, p.Server.Addr, api.HealthCheckInterval, api.HealthCheckTimeout); err != nil {
							return err
						}

						return nil
					},
					OnStop: func(ctx context.Context) error {
						state.mu.Lock()
						if !state.started || state.shutdown {
							state.mu.Unlock()
							return nil
						}
						state.shutdown = true
						state.mu.Unlock()

						// Wait for server goroutine to exit
						select {
						case <-state.serverDone:
							p.Logger.Info("HTTP server goroutine exited")
						case <-ctx.Done():
							p.Logger.Warn("Timeout waiting for server goroutine to exit")
						}

						return Shutdown(ctx, p.Server)
					},
				})
			}),
		)

		// Set the fx app for coordinated shutdown
		handler.SetFXApp(fxApp)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for shutdown signal
		handler.Wait()

		return nil
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
