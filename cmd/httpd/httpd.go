// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/constants"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
)

// Cmd represents the HTTP server command
var Cmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Get dependencies
		deps, err := cmdcommon.NewCommandDeps()
		if err != nil {
			return fmt.Errorf("failed to initialize dependencies: %w", err)
		}

		// Create storage using common function
		storageResult, err := cmdcommon.CreateStorage(deps.Config, deps.Logger)
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
			shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.DefaultShutdownTimeout)
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
