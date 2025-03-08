// Package cmd implements the command-line interface for GoCrawl.
// This file contains the HTTP server command implementation that provides a REST API
// for searching content in Elasticsearch.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// ServerParams holds the parameters required for running the HTTP server.
// It uses fx.In for dependency injection of required components.
type ServerParams struct {
	fx.In

	// Server is the HTTP server instance that handles incoming requests
	Server *http.Server
	// Logger provides logging capabilities for the HTTP server
	Logger logger.Interface
}

// httpdCmd represents the HTTP server command that provides a REST API
// for searching content in Elasticsearch.
var httpdCmd = &cobra.Command{
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
			api.Module,
			fx.Invoke(func(lc fx.Lifecycle, p ServerParams) {
				lc.Append(fx.Hook{
					OnStart: func(context.Context) error {
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
		case <-cmd.Context().Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
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
			if serverErr == nil {
				serverErr = err
			}
		}

		return serverErr
	},
}

// init initializes the HTTP server command by adding it to the root command.
func init() {
	rootCmd.AddCommand(httpdCmd)
}
