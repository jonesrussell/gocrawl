// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"go.uber.org/fx"
)

const (
	// healthCheckTimeout is the maximum time to wait for the server to become healthy
	healthCheckTimeout = 5 * time.Second
	// healthCheckInterval is the time between health check attempts
	healthCheckInterval = 100 * time.Millisecond
)

// SearchRequest represents the structure of the search request
type SearchRequest struct {
	Query string `json:"query"`
	Index string `json:"index"`
	Size  int    `json:"size"`
}

// SearchResponse represents the structure of the search response
type SearchResponse struct {
	Results []any `json:"results"`
	Total   int   `json:"total"`
}

// Module provides API dependencies
var Module = fx.Module("api",
	common.Module,
	fx.Provide(
		// Provide the server and security middleware together to avoid circular dependencies
		func(
			log common.Logger,
			searchManager SearchManager,
			cfg common.Config,
			lc fx.Lifecycle,
		) (*http.Server, middleware.SecurityMiddlewareInterface) {
			// Create router and security middleware
			router, security := SetupRouter(log, searchManager, cfg)

			// Create server
			server := &http.Server{
				Addr:              cfg.GetServerConfig().Address,
				Handler:           router,
				ReadHeaderTimeout: readHeaderTimeout,
			}

			// Register lifecycle hooks
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Create a channel to signal when the server is ready
					ready := make(chan struct{})
					go func() {
						// Start the server
						if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							log.Error("Server error", "error", err)
						}
					}()

					// Create a timeout context for health check
					healthCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
					defer cancel()

					// Create a ticker for health check attempts
					ticker := time.NewTicker(healthCheckInterval)
					defer ticker.Stop()

					// Try to connect to the health endpoint until successful or timeout
					for {
						select {
						case <-healthCtx.Done():
							return fmt.Errorf("server failed to become healthy within %v", healthCheckTimeout)
						case <-ticker.C:
							// Create a temporary client for health check
							client := &http.Client{
								Timeout: healthCheckInterval,
							}

							// Try to connect to the health endpoint
							resp, err := client.Get(fmt.Sprintf("http://%s/health", server.Addr))
							if err != nil {
								continue // Server not ready yet
							}
							resp.Body.Close()

							if resp.StatusCode == http.StatusOK {
								close(ready)
								return nil
							}
						}
					}
				},
				OnStop: func(ctx context.Context) error {
					// Create a timeout context for shutdown
					shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
					defer cancel()

					// Shutdown server
					if err := server.Shutdown(shutdownCtx); err != nil {
						return fmt.Errorf("server shutdown failed: %w", err)
					}

					// Cleanup security middleware
					security.Cleanup(ctx)
					security.WaitCleanup()

					// Close search manager
					if err := searchManager.Close(); err != nil {
						return fmt.Errorf("search manager close failed: %w", err)
					}

					return nil
				},
			})

			return server, security
		},
	),
	fx.Invoke(ConfigureLifecycle),
)
