// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
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
		// Get dependencies from context using helper
		log, cfg, err := cmdcommon.GetDependencies(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get dependencies: %w", err)
		}

		// Construct dependencies directly without FX
		storageClient, err := createStorageClientForHttpd(cfg, log)
		if err != nil {
			return fmt.Errorf("failed to create storage client: %w", err)
		}

		storageResult, err := storage.NewStorage(storage.StorageParams{
			Config: cfg,
			Logger: log,
			Client: storageClient,
		})
		if err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}

		// Create search manager
		searchManager := storage.NewSearchManager(storageResult.Storage, log)

		// Create HTTP server
		srv, _, err := api.StartHTTPServer(log, searchManager, cfg)
		if err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		// Start server in goroutine
		log.Info("Starting HTTP server", "addr", cfg.GetServerConfig().Address)
		errChan := make(chan error, 1)
		go func() {
			if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
				errChan <- serveErr
			}
		}()

		// Wait for interrupt signal or error
		select {
		case serverErr := <-errChan:
			log.Error("Server error", "error", serverErr)
			return fmt.Errorf("server error: %w", serverErr)
		case <-cmd.Context().Done():
			// Graceful shutdown with timeout
			log.Info("Shutdown signal received")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cmdcommon.DefaultShutdownTimeout)
			defer cancel()

			log.Info("Stopping HTTP server")
			if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
				log.Error("Failed to stop server", "error", shutdownErr)
				return fmt.Errorf("failed to stop server: %w", shutdownErr)
			}

			log.Info("Server stopped successfully")
			return nil
		}
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}

// createStorageClientForHttpd creates an Elasticsearch client for the httpd command
func createStorageClientForHttpd(cfg config.Interface, log logger.Interface) (*es.Client, error) {
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: cfg,
		Logger: log,
	})
	if err != nil {
		return nil, err
	}
	return clientResult.Client, nil
}
