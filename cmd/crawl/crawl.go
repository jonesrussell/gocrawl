// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Cmd represents the crawl command
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
You can specify the source to crawl using the --source flag.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set the source flag from the argument
		if err := cmd.Flags().Set("source", args[0]); err != nil {
			return fmt.Errorf("failed to set source flag: %w", err)
		}

		// Set debug mode if enabled
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			if err := os.Setenv("APP_DEBUG", "true"); err != nil {
				return fmt.Errorf("failed to set debug environment variable: %w", err)
			}
			if err := os.Setenv("LOG_DEBUG", "true"); err != nil {
				return fmt.Errorf("failed to set debug environment variable: %w", err)
			}
		}

		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling with a no-op logger initially
		handler := signal.NewSignalHandler(logger.NewNoOp())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Initialize the Fx application
		fxApp := fx.New(
			crawler.Module,
			storage.Module,
			sources.Module,
			fx.Provide(
				fx.Annotate(
					func() context.Context { return ctx },
					fx.ResultTags(`name:"crawlContext"`),
				),
				fx.Annotate(
					func() string {
						sourceName, _ := cmd.Flags().GetString("source")
						return sourceName
					},
					fx.ResultTags(`name:"sourceName"`),
				),
				fx.Annotate(
					func() *signal.SignalHandler { return handler },
					fx.ResultTags(`name:"signalHandler"`),
				),
				fx.Annotate(
					func() chan struct{} { return make(chan struct{}) },
					fx.ResultTags(`name:"shutdownChan"`),
				),
			),
			fx.Invoke(func(lc fx.Lifecycle, p CommandDeps) {
				// Update the signal handler with the real logger
				handler.SetLogger(p.Logger)
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test storage connection
						if err := p.Storage.TestConnection(ctx); err != nil {
							return fmt.Errorf("failed to connect to storage: %w", err)
						}

						// Start crawler
						p.Logger.Info("Starting crawler...", "source", p.SourceName)
						if err := p.Crawler.Start(ctx, p.SourceName); err != nil {
							return fmt.Errorf("failed to start crawler: %w", err)
						}

						// Wait for crawler to complete
						go func() {
							p.Crawler.Wait()
							p.Logger.Info("Crawler finished processing")
							// Signal completion to the signal handler
							handler.RequestShutdown()
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						// Stop crawler
						if err := p.Crawler.Stop(ctx); err != nil {
							return fmt.Errorf("failed to stop crawler: %w", err)
						}

						// Close storage connection
						if err := p.Storage.Close(); err != nil {
							return fmt.Errorf("failed to close storage connection: %w", err)
						}

						return nil
					},
				})
			}),
		)

		// Start the application
		if err := fxApp.Start(context.Background()); err != nil {
			return fmt.Errorf("failed to start application: %w", err)
		}

		// Wait for completion signal
		handler.Wait()

		return nil
	},
}

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}

func init() {
	Cmd.Flags().String("source", "", "The source to crawl")
	Cmd.Flags().Bool("debug", false, "Enable debug mode")
}
