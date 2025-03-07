package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Initialize the Fx application with the HTTP server
		app := fx.New(
			fx.Provide(
				func() *config.Config {
					return globalConfig // Provide the global config
				},
				func() logger.Interface {
					return globalLogger // Use the global logger
				},
			),
			api.Module,
			storage.Module,
			fx.Invoke(func(server *http.Server) {
				// Start the server
				go func() {
					if err := server.ListenAndServe(); err != nil {
						globalLogger.Error("HTTP server failed", "error", err)
					}
				}()
				globalLogger.Info("HTTP server started on :8081")
			}),
		)

		// Start the application
		if err := app.Start(context.Background()); err != nil {
			globalLogger.Error("Error starting application", "error", err)
			return
		}

		// Wait for a termination signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan // Block until a signal is received

		// Wait for the application to stop
		defer func() {
			if err := app.Stop(context.Background()); err != nil {
				globalLogger.Error("Error stopping application", "error", err)
			}
		}()
		globalLogger.Debug("HTTP server stopped")
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)
}
