// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// CommandDeps represents the dependencies required by the crawl command.
type CommandDeps struct {
	fx.In

	Context     context.Context `name:"crawlContext"`
	SourceName  string          `name:"sourceName"`
	Logger      logger.Interface
	Crawler     crawler.Interface
	Sources     *sources.Sources
	Handler     *signal.SignalHandler
	ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
	Processors  []common.Processor   `group:"processors"`
}

// CrawlerCommand represents the crawl command.
type CrawlerCommand struct {
	crawler     crawler.Interface
	logger      logger.Interface
	processors  []common.Processor
	articleChan chan *models.Article
	done        chan struct{}
}

// NewCrawlerCommand creates a new crawl command.
func NewCrawlerCommand(
	logger logger.Interface,
	crawler crawler.Interface,
	processors []common.Processor,
	articleChan chan *models.Article,
) *CrawlerCommand {
	return &CrawlerCommand{
		crawler:     crawler,
		logger:      logger,
		processors:  processors,
		articleChan: articleChan,
		done:        make(chan struct{}),
	}
}

// Start starts the crawler.
func (c *CrawlerCommand) Start(ctx context.Context, sourceName string) error {
	// Start crawler
	c.logger.Info("Starting crawler...", "source", sourceName)
	if err := c.crawler.Start(ctx, sourceName); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Process articles from the crawler's channel
	go func() {
		defer close(c.done)
		crawlerChan := c.crawler.GetArticleChannel()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping article processing")
				return
			case article, ok := <-crawlerChan:
				if !ok {
					c.logger.Info("Crawler article channel closed")
					return
				}

				// Forward article to the command's channel
				select {
				case c.articleChan <- article:
					// Article forwarded successfully
				case <-ctx.Done():
					c.logger.Info("Context cancelled while forwarding article")
					return
				}

				// Create a job for each article with timeout
				job := &common.Job{
					ID:        article.ID,
					URL:       article.Source,
					Status:    "pending",
					CreatedAt: article.CreatedAt,
					UpdatedAt: article.UpdatedAt,
				}

				// Process the article using each processor with timeout
				for _, processor := range c.processors {
					// Create a timeout context for each processor
					processorCtx, processorCancel := context.WithTimeout(ctx, 30*time.Second)
					processor.ProcessJob(processorCtx, job)
					processorCancel()
				}
			}
		}
	}()

	return nil
}

// Stop stops the crawler.
func (c *CrawlerCommand) Stop(ctx context.Context) error {
	return c.crawler.Stop(ctx)
}

// Wait waits for the crawler to finish processing.
func (c *CrawlerCommand) Wait() {
	c.crawler.Wait()
	<-c.done
}

// Module provides the crawl command module for dependency injection.
var Module = fx.Module("crawl",
	// Include all required modules
	config.Module,
	storage.Module,
	logger.Module,
	crawler.Module,
	sources.Module,
	article.Module,
	content.Module,
	// Provide context and source name
	fx.Provide(
		func() context.Context { return context.Background() },
		func() string { return "" }, // Will be overridden by command
	),
	// Provide logger params
	fx.Provide(func() logger.Params {
		return logger.Params{
			Config: &logger.Config{
				Level:       logger.InfoLevel,
				Development: true,
				Encoding:    "console",
			},
		}
	}),
	// Provide article channel
	fx.Provide(func() chan *models.Article {
		return make(chan *models.Article, 100)
	}),
	// Provide processors
	fx.Provide(
		// Provide article processor
		func(
			logger logger.Interface,
			config config.Interface,
			storage storagetypes.Interface,
			service article.Interface,
		) *article.ArticleProcessor {
			return article.ProvideArticleProcessor(logger, config, storage, service)
		},
		// Provide content processor
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
		// Provide processor slice
		func(
			articleProcessor *article.ArticleProcessor,
			contentProcessor *content.ContentProcessor,
		) []common.Processor {
			return []common.Processor{
				articleProcessor,
				contentProcessor,
			}
		},
	),
	// Provide the event bus
	fx.Provide(events.NewBus),
	// Provide the IndexManager
	fx.Provide(func(client *elasticsearch.Client, logger logger.Interface) interfaces.IndexManager {
		return storage.NewElasticsearchIndexManager(client, logger)
	}),
	// Provide signal handler
	fx.Provide(func(ctx context.Context, logger logger.Interface) *signal.SignalHandler {
		handler := signal.NewSignalHandler(logger)
		handler.Setup(ctx)
		return handler
	}),
	// Invoke the crawler lifecycle
	fx.Invoke(func(lc fx.Lifecycle, deps CommandDeps) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Start the crawler
				if err := deps.Crawler.Start(ctx, deps.SourceName); err != nil {
					return fmt.Errorf("failed to start crawler: %w", err)
				}

				// Start a goroutine to wait for crawler completion
				go func() {
					// Create a timeout context for waiting
					waitCtx, waitCancel := context.WithTimeout(ctx, crawlerTimeout)
					defer waitCancel()

					// Wait for crawler to complete
					deps.Crawler.Wait()

					// Check if we timed out
					select {
					case <-waitCtx.Done():
						deps.Logger.Info("Crawler reached timeout limit")
					default:
						deps.Logger.Info("Crawler finished processing")
					}

					// Signal completion to the signal handler
					deps.Handler.RequestShutdown()
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				// Stop the crawler with timeout
				stopCtx, stopCancel := context.WithTimeout(ctx, shutdownTimeout)
				defer stopCancel()

				if err := deps.Crawler.Stop(stopCtx); err != nil {
					return fmt.Errorf("failed to stop crawler: %w", err)
				}
				return nil
			},
		})
	}),
)
