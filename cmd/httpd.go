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

// ServerParams holds the parameters for the HTTP server
type ServerParams struct {
	fx.In

	Server *http.Server
	Logger common.Logger
}

// startServer starts the HTTP server and handles graceful shutdown
func startServer(lc fx.Lifecycle, p ServerParams) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				p.Logger.Info("HTTP server started", "address", p.Server.Addr)
				if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("HTTP server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Shutting down HTTP server...")
			return p.Server.Shutdown(ctx)
		},
	})
}

// httpdCmd represents the httpd command
var httpdCmd = &cobra.Command{
	Use:   "httpd",
	Short: "Start the HTTP server for search",
	Long: `This command starts an HTTP server that listens for search requests.
You can send POST requests to /search with a JSON body containing the search parameters.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Initialize the Fx application with the HTTP server
		app := fx.New(
			common.Module,
			api.Module,
			fx.Invoke(startServer),
		)

		// Start the application
		if err := app.Start(context.Background()); err != nil {
			common.PrintErrorf("Error starting application: %v", err)
			os.Exit(1)
		}

		// Wait for termination signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
		defer func() {
			cancel()
			if err := app.Stop(ctx); err != nil {
				common.PrintErrorf("Error during shutdown: %v", err)
				os.Exit(1)
			}
		}()
	},
}

func init() {
	rootCmd.AddCommand(httpdCmd)
}
