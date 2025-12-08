// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
)

// LifecycleParams holds the dependencies for SetupLifecycle
type LifecycleParams struct {
	LC            fx.Lifecycle // Keep for now if ConfigureLifecycle needs it
	Ctx           context.Context
	Server        *http.Server
	SearchManager SearchManager
	Security      middleware.SecurityMiddlewareInterface
	Log           logger.Interface
}

// Lifecycle manages the API server lifecycle
type Lifecycle struct {
	server *http.Server
	Log    logger.Interface
}

// NewLifecycle creates a new API lifecycle manager
func NewLifecycle(server *http.Server, log logger.Interface) *Lifecycle {
	return &Lifecycle{
		server: server,
		Log:    log,
	}
}

// Start starts the API server
func (l *Lifecycle) Start(ctx context.Context) error {
	l.Log.Info("Starting API server", "addr", l.server.Addr)
	return l.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (l *Lifecycle) Stop(ctx context.Context) error {
	l.Log.Info("Stopping API server")
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()
	return l.server.Shutdown(shutdownCtx)
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
			shutdownCtx, shutdownCancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
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

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context, app *fx.App) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	return app.Stop(shutdownCtx)
}
