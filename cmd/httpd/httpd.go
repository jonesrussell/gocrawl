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

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// maxRetries is the maximum number of retries for Elasticsearch connection
	maxRetries = 3
	// retryDelay is the delay between retries
	retryDelay = 5 * time.Second
)

// Params holds the parameters required for running the HTTP server.
// It uses fx.In for dependency injection of required components.
type Params struct {
	fx.In

	// Server is the HTTP server instance that handles incoming requests
	Server *http.Server
	// Logger provides logging capabilities for the HTTP server
	Logger logger.Interface
	// Storage provides access to Elasticsearch
	Storage storage.Interface
}

// Cmd represents the HTTP server command that provides a REST API
// for searching content in Elasticsearch.
var Cmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create channels for error handling and completion
		errChan := make(chan error, 1)
		doneChan := make(chan struct{})

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			Module,
			fx.Invoke(func(lc fx.Lifecycle, p Params) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test Elasticsearch connection with retries
						for i := range maxRetries {
							// Try to test the connection
							err := p.Storage.TestConnection(ctx)
							if err == nil {
								break
							}
							if i < maxRetries-1 {
								p.Logger.Warn("Failed to connect to Elasticsearch, retrying...", "error", err, "attempt", i+1)
								time.Sleep(retryDelay)
								continue
							}
							return fmt.Errorf("failed to connect to Elasticsearch after %d attempts: %w", maxRetries, err)
						}

						// Start the server in a goroutine
						go func() {
							p.Logger.Info("HTTP server started", "address", p.Server.Addr)
							if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								p.Logger.Error("HTTP server failed", "error", err)
								errChan <- err
								return
							}
							close(doneChan)
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						p.Logger.Info("Shutting down HTTP server...")
						return p.Server.Shutdown(ctx)
					},
				})
			}),
		)

		// Start the application and handle any startup errors
		if err := app.Start(cmd.Context()); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either:
		// - A signal interrupt (SIGINT/SIGTERM)
		// - Context cancellation
		// - Server error
		// - Server shutdown
		var serverErr error
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
			// Graceful shutdown requested, not an error
		case <-cmd.Context().Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
			// Context cancellation requested, not an error
		case serverErr = <-errChan:
			// Error already logged in OnStart hook
		case <-doneChan:
			// Server shut down cleanly
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(cmd.Context(), common.DefaultShutdownTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error during shutdown: %v", err)
			if serverErr == nil && !errors.Is(err, context.Canceled) {
				serverErr = err
			}
		}

		// Only return error for actual failures, not graceful shutdowns
		if serverErr != nil && !errors.Is(serverErr, context.Canceled) {
			return serverErr
		}
		return nil
	},
}

// Command returns the httpd command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
