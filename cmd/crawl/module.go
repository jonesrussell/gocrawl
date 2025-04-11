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
				return make(chan *models.Article, 100)
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
		signal.NewSignalHandler,

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
		func(lc fx.Lifecycle, logger logger.Interface, crawler crawler.Interface, handler signal.Interface, sourceName string) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Set up signal handling
					handler.Setup(ctx)

					// Start the crawler
					if err := crawler.Start(ctx, sourceName); err != nil {
						return fmt.Errorf("failed to start crawler: %w", err)
					}

					// Start a goroutine to wait for crawler completion
					go func() {
						// Create a timeout context for waiting
						waitCtx, waitCancel := context.WithTimeout(ctx, crawlerTimeout)
						defer waitCancel()

						// Wait for crawler to complete
						crawler.Wait()

						// Check if we timed out
						select {
						case <-waitCtx.Done():
							logger.Info("Crawler reached timeout limit")
						default:
							logger.Info("Crawler finished processing")
						}

						// Signal completion to the signal handler
						handler.RequestShutdown()
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Stop the crawler with timeout
					stopCtx, stopCancel := context.WithTimeout(ctx, shutdownTimeout)
					defer stopCancel()

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
