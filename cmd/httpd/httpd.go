// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful server shutdown.
	DefaultShutdownTimeout = 5 * time.Second
)

// Params holds the parameters required for running the HTTP server.
type Params struct {
	Server *http.Server
	Logger logger.Interface
	Config config.Interface
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
		// Get dependencies - NEW WAY
		deps, err := cmdcommon.NewCommandDeps()
		if err != nil {
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		// Create storage using common function
		storageClient, err := cmdcommon.CreateStorageClient(deps.Config, deps.Logger)
		if err != nil {
			return fmt.Errorf("failed to create storage client: %w", err)
		}

		storageResult, err := storage.NewStorage(storage.StorageParams{
			Config: deps.Config,
			Logger: deps.Logger,
			Client: storageClient,
		})
		if err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}

		// Create search manager
		searchManager := storage.NewSearchManager(storageResult.Storage, deps.Logger)

		// Create HTTP server
		srv, _, err := api.StartHTTPServer(deps.Logger, searchManager, deps.Config)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		// Start server in goroutine
		deps.Logger.Info("Starting HTTP server", "addr", deps.Config.GetServerConfig().Address)
		errChan := make(chan error, 1)
		go func() {
			if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
				errChan <- serveErr
			}
		}()

		// Wait for interrupt signal or error
		select {
		case serverErr := <-errChan:
			deps.Logger.Error("Server error", "error", serverErr)
			return fmt.Errorf("server error: %w", serverErr)
		case <-cmd.Context().Done():
			// Graceful shutdown with timeout
			deps.Logger.Info("Shutdown signal received")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cmdcommon.DefaultShutdownTimeout)
			defer cancel()

			deps.Logger.Info("Stopping HTTP server")
			if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
				deps.Logger.Error("Failed to stop server", "error", shutdownErr)
				return fmt.Errorf("failed to stop server: %w", shutdownErr)
			}

			deps.Logger.Info("Server stopped successfully")
			return nil
		}
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
