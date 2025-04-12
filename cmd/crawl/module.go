// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
)

// Module provides the crawl command module for dependency injection.
var Module = fx.Options(
	// Include all required modules
	config.Module,
	storage.Module,
	logger.Module,
	sources.Module,
	article.Module,
	content.Module,

	// Provide base dependencies
	fx.Provide(
		// Logger params
		func() logger.Params {
			return logger.Params{
				Config: &logger.Config{
					Level:       logger.InfoLevel,
					Development: true,
					Encoding:    "console",
				},
			}
		},

		// Article channel
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),

		// Sources
		func(config config.Interface, logger logger.Interface) (*sources.Sources, error) {
			return sources.LoadSources(config)
		},

		// Event bus
		events.NewEventBus,

		// Index manager
		storage.NewElasticsearchIndexManager,

		// Signal handler
		fx.Annotate(
			signal.NewSignalHandler,
			fx.As(new(signal.Interface)),
		),

		// Article processor
		article.ProvideArticleProcessor,

		// Content processor
		func(
			logger logger.Interface,
			service content.Interface,
			storage storagetypes.Interface,
		) *content.ContentProcessor {
			return content.NewContentProcessor(content.ProcessorParams{
				Logger:    logger,
				Service:   service,
				Storage:   storage,
				IndexName: "content",
			})
		},

		// Processors slice
		func(articleProcessor *article.ArticleProcessor, contentProcessor *content.ContentProcessor) []common.Processor {
			return []common.Processor{articleProcessor, contentProcessor}
		},
	),

	// Include crawler module after processors are provided
	crawler.Module,

	// Invoke the crawler lifecycle
	fx.Invoke(fx.Annotate(
		func(
			lc fx.Lifecycle,
			logger logger.Interface,
			crawler crawler.Interface,
			handler signal.Interface,
			sourceName string,
		) {
			// Set up signal handling
			cleanup := handler.Setup(context.Background())
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					cleanup()
					return nil
				},
			})

			// Set up crawler lifecycle
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Create a context with timeout for the crawler
					crawlCtx, cancel := context.WithTimeout(ctx, common.DefaultOperationTimeout)
					defer cancel()

					// Start the crawler
					if err := crawler.Start(crawlCtx, sourceName); err != nil {
						return fmt.Errorf("failed to start crawler: %w", err)
					}

					// Start a goroutine to wait for crawler completion
					go func() {
						// Wait for crawler to complete
						select {
						case <-crawler.Done():
							logger.Info("Crawler finished processing")
						case <-crawlCtx.Done():
							logger.Info("Crawler context cancelled")
						}

						// Signal completion to the signal handler
						handler.RequestShutdown()
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Create a context with timeout for stopping the crawler
					stopCtx, cancel := context.WithTimeout(ctx, common.DefaultOperationTimeout)
					defer cancel()

					// Stop the crawler
					if err := crawler.Stop(stopCtx); err != nil {
						return fmt.Errorf("failed to stop crawler: %w", err)
					}
					return nil
				},
			})
		},
		fx.ParamTags(``, ``, ``, ``, `name:"sourceName"`),
	)),
)
