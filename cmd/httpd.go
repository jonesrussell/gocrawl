// Package cmd implements the command-line interface for GoCrawl.
// This file contains the HTTP server command implementation that provides a REST API
// for searching content in Elasticsearch.
package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
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
	Logger common.Logger
}

// startServer initializes and manages the HTTP server lifecycle.
// It:
// - Starts the server in a goroutine to handle incoming requests
// - Handles graceful shutdown when the application stops
// - Logs server events and errors
func startServer(lc fx.Lifecycle, p ServerParams) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Start the server in a goroutine to allow for graceful shutdown
			go func() {
				p.Logger.Info("HTTP server started", "address", p.Server.Addr)
				if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("HTTP server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Gracefully shut down the server when the application stops
			p.Logger.Info("Shutting down HTTP server...")
			return p.Server.Shutdown(ctx)
		},
	})
}

// httpdCmd represents the HTTP server command that provides a REST API
// for searching content in Elasticsearch.
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			api.Module,
			fx.Invoke(startServer),
		)

		// Start the application and handle any startup errors
		if err := app.Start(cmd.Context()); err != nil {
			common.PrintErrorf("Error starting application: %v", err)
			os.Exit(1)
		}

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultShutdownTimeout)
		defer func() {
			cancel()
			if err := app.Stop(ctx); err != nil {
				common.PrintErrorf("Error during shutdown: %v", err)
				os.Exit(1)
			}
		}()
	},
}

// init initializes the HTTP server command by adding it to the root command.
func init() {
	rootCmd.AddCommand(httpdCmd)
}
