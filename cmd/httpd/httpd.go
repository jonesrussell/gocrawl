// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

// Params holds the parameters required for running the HTTP server.
type Params struct {
	fx.In
	Server  *http.Server
	Logger  logger.Interface
	Storage storage.Interface
}

// serverState tracks the HTTP server's state
type serverState struct {
	mu       sync.Mutex
	started  bool
	shutdown bool
	// serverDone is closed when the server goroutine exits
	serverDone chan struct{}
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
		handler := signalhandler.NewSignalHandler()
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Track server state
		state := &serverState{
			serverDone: make(chan struct{}),
		}

		// Initialize the Fx application
		fxApp := fx.New(
			common.Module,
			Module,
			fx.Provide(
				func() context.Context { return ctx },
			),
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
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

						go func() {
							defer close(state.serverDone)
							if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								p.Logger.Error("HTTP server failed", "error", err)
							}
						}()

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

						p.Logger.Info("Initiating graceful shutdown...")

						// Create timeout context for shutdown
						shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
						defer cancel()

						// Shutdown HTTP server
						p.Logger.Info("Shutting down HTTP server...")
						if err := p.Server.Shutdown(shutdownCtx); err != nil {
							if err != context.Canceled && err != context.DeadlineExceeded {
								p.Logger.Error("Error during server shutdown", "error", err)
								return fmt.Errorf("error during server shutdown: %w", err)
							}
						}

						// Wait for server goroutine to exit
						select {
						case <-state.serverDone:
							p.Logger.Info("HTTP server goroutine exited")
						case <-shutdownCtx.Done():
							p.Logger.Warn("Timeout waiting for server goroutine to exit")
						}

						// Close storage connection
						p.Logger.Info("Closing storage connection...")
						if err := p.Storage.Close(); err != nil {
							p.Logger.Error("Error closing storage connection", "error", err)
							// Don't return error here as server is already stopped
						}

						p.Logger.Info("Shutdown complete")
						return nil
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
