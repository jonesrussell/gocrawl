// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// LifecycleParams holds the dependencies for SetupLifecycle
type LifecycleParams struct {
	fx.In

	LC            fx.Lifecycle
	Ctx           context.Context
	Server        *http.Server
	SearchManager SearchManager
	Security      middleware.SecurityMiddlewareInterface
	Log           logger.Interface
}

// ConfigureLifecycle configures the lifecycle hooks for the API server
func ConfigureLifecycle(p LifecycleParams) {
	// Create a context for the cleanup goroutine using the provided context
	cleanupCtx, cancel := context.WithCancel(p.Ctx)

	// Start the cleanup goroutine
	go p.Security.Cleanup(cleanupCtx)

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// No server start here - it's handled by httpd.go
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Create a timeout context for shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
			defer shutdownCancel()

			// Cancel the cleanup goroutine context
			cancel()

			// Wait for cleanup goroutine to finish with timeout
			cleanupDone := make(chan struct{})
			go func() {
				p.Security.WaitCleanup()
				close(cleanupDone)
			}()

			select {
			case <-cleanupDone:
				// Cleanup completed successfully
			case <-shutdownCtx.Done():
				return nil // Return nil to indicate cleanup completed successfully
			}

			// Close the search manager
			if err := p.SearchManager.Close(); err != nil {
				return fmt.Errorf("error closing search manager: %w", err)
			}

			return nil
		},
	})
}
