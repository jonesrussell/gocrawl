// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Dependencies holds the crawl command's dependencies
type Dependencies struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Sources     sources.Interface `name:"sourceManager"`
	Crawler     crawler.Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     types.Interface
	Done        chan struct{}             `name:"crawlDone"`
	Context     context.Context           `name:"crawlContext"`
	Processors  []models.ContentProcessor `group:"processors"`
	SourceName  string                    `name:"sourceName"`
	ArticleChan chan *models.Article      `name:"articleChannel"`
}

// Cmd represents the crawl command
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
You can specify the source to crawl using the --source flag.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling with a no-op logger initially
		handler := signal.NewSignalHandler(logger.NewNoOp())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Initialize the Fx application
		fxApp := fx.New(
			fx.NopLogger,
			common.Module,
			crawler.Module,
			fx.Provide(
				func() context.Context { return ctx },
			),
			fx.Invoke(func(lc fx.Lifecycle, p Dependencies) {
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

						return nil
					},
					OnStop: func(ctx context.Context) error {
						p.Logger.Info("Stopping crawler...")
						if err := p.Crawler.Stop(ctx); err != nil {
							return fmt.Errorf("failed to stop crawler: %w", err)
						}

						// Close storage connection
						p.Logger.Info("Closing storage connection...")
						if err := p.Storage.Close(); err != nil {
							p.Logger.Error("Error closing storage connection", "error", err)
							// Don't return error here as crawler is already stopped
						}

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

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
