// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// sourceName holds the name of the source being crawled, populated from command line arguments.
var sourceName string

// SetSourceName sets the source name for testing purposes.
func SetSourceName(name string) {
	sourceName = name
}

// Params holds the dependencies and parameters required for the crawl operation.
type Params struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Sources    sources.Interface `name:"sourceManager"`
	Crawler    crawler.Interface
	Logger     logger.Interface
	Config     config.Interface
	Done       chan struct{}             `name:"crawlDone"`
	Context    context.Context           `name:"crawlContext"`
	Processors []models.ContentProcessor `group:"processors"`
}

// StartCrawl initializes and starts the crawling process.
func StartCrawl(p Params) error {
	source, findErr := p.Sources.FindByName(sourceName)
	if findErr != nil {
		return fmt.Errorf("error finding source: %w", findErr)
	}

	// Register lifecycle hooks
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			collectorResult, setupErr := app.SetupCollector(ctx, p.Logger, *source, p.Processors, p.Done, p.Config)
			if setupErr != nil {
				return fmt.Errorf("error setting up collector: %w", setupErr)
			}

			if configErr := app.ConfigureCrawler(p.Crawler, *source, collectorResult); configErr != nil {
				return fmt.Errorf("error configuring crawler: %w", configErr)
			}

			p.Logger.Info("Starting crawl", "source", source.Name)
			if startErr := p.Crawler.Start(ctx, source.URL); startErr != nil {
				return fmt.Errorf("error starting crawler: %w", startErr)
			}

			// Monitor crawl completion in a separate goroutine
			go func() {
				p.Crawler.Wait()
				p.Logger.Info("Crawler finished processing all URLs")

				// Signal completion through the Done channel
				close(p.Done)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping crawler")
			if stopErr := p.Crawler.Stop(ctx); stopErr != nil {
				return fmt.Errorf("error stopping crawler: %w", stopErr)
			}
			return nil
		},
	})

	return nil
}

// Cmd represents the crawl command.
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a single source defined in sources.yml",
	Long: `Crawl a single source defined in sources.yml.
The source argument must match a name defined in your sources.yml configuration file.

Example:
  gocrawl crawl example-blog`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			const errMsg = "requires exactly one source name from sources.yml\n\n" +
				"Usage:\n  %s\n\n" +
				"Run 'gocrawl list' to see available sources"
			return fmt.Errorf(errMsg, cmd.UseLine())
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName = args[0]

		// Create a cancellable context and set up signal handling
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		signalDone, cleanup := app.WaitForSignal(ctx, cancel)
		defer cleanup()

		// Create a logger for the command
		log, logErr := zap.NewProduction()
		if logErr != nil {
			return fmt.Errorf("error creating logger: %w", logErr)
		}
		defer func() {
			if syncErr := log.Sync(); syncErr != nil {
				// We can't return an error here since we're in a defer,
				// but we can at least try to log it
				log.Error("Failed to sync logger", zap.Error(syncErr))
			}
		}()

		// Create the crawler's completion channel
		crawlerDone := make(chan struct{})

		// Initialize the Fx application with required modules
		fxApp := fx.New(
			fx.NopLogger,
			// Core dependencies
			config.Module,
			logger.Module,
			storage.Module,
			sources.Module,
			api.Module,

			// Feature modules
			article.Module,
			content.Module,
			collector.Module(),
			crawler.Module,

			fx.Provide(
				fx.Annotate(
					func() chan struct{} { return crawlerDone },
					fx.ResultTags(`name:"crawlDone"`),
				),
				fx.Annotate(
					func() context.Context { return ctx },
					fx.ResultTags(`name:"crawlContext"`),
				),
				fx.Annotate(
					func() string { return sourceName },
					fx.ResultTags(`name:"sourceName"`),
				),
				fx.Annotate(
					func(sources sources.Interface) (string, string) {
						src, srcErr := sources.FindByName(sourceName)
						if srcErr != nil {
							return "", ""
						}
						return src.Index, src.ArticleIndex
					},
					fx.ParamTags(`name:"sourceManager"`),
					fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
				),
				func() chan *models.Article {
					return make(chan *models.Article, app.DefaultChannelBufferSize)
				},
			),
			fx.Invoke(StartCrawl),
		)

		// Start the application
		if startErr := fxApp.Start(ctx); startErr != nil {
			if errors.Is(startErr, context.Canceled) {
				return nil
			}
			return fmt.Errorf("error starting application: %w", startErr)
		}

		// Wait for completion or cancellation
		select {
		case <-signalDone:
			log.Info("Received interrupt signal")
		case <-crawlerDone:
			log.Info("Crawler completed successfully")
		case <-ctx.Done():
			log.Info("Context cancelled")
		}

		// Perform graceful shutdown
		if shutdownErr := app.GracefulShutdown(fxApp); shutdownErr != nil {
			log.Error("Error during shutdown", zap.Error(shutdownErr))
		}

		return nil
	},
}

// Command returns the crawl command.
func Command() *cobra.Command {
	return Cmd
}
