// Package crawler provides the core crawling functionality for GoCrawl.
// It manages the crawling process, including URL processing, rate limiting,
// and content extraction.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
)

// Interface defines the crawler's capabilities.
type Interface interface {
	// Start begins crawling from the given base URL.
	Start(ctx context.Context, sourceName string) error
	// Stop gracefully stops the crawler.
	Stop(ctx context.Context) error
	// Subscribe adds a content handler to receive discovered content.
	Subscribe(handler events.Handler)
	// SetRateLimit sets the crawler's rate limit.
	SetRateLimit(duration time.Duration) error
	// SetMaxDepth sets the maximum crawl depth.
	SetMaxDepth(depth int)
	// SetCollector sets the collector for the crawler.
	SetCollector(collector *colly.Collector)
	// GetIndexManager returns the index manager interface.
	GetIndexManager() api.IndexManager
	// Wait blocks until the crawler has finished processing all queued requests.
	Wait()
	// GetMetrics returns the current crawler metrics.
	GetMetrics() *collector.Metrics
}

// Result holds the crawler module's output.
type Result struct {
	fx.Out
	Crawler Interface
}

// Module provides the crawler module's dependencies.
var Module = fx.Module("crawler",
	// Include core dependencies
	config.Module,
	logger.Module,
	sources.Module,
	fx.Provide(
		// Provide core dependencies
		func() context.Context {
			return context.Background()
		},
		// Provide named dependencies
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"crawlerDoneChannel"`),
		),
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
		fx.Annotate(
			func() string {
				return "articles"
			},
			fx.ResultTags(`name:"indexName"`),
		),
		// Provide article service
		fx.Annotate(
			func(logger logger.Interface, storage storagetypes.Interface) article.Interface {
				return article.NewService(logger, config.DefaultArticleSelectors(), storage, "articles")
			},
		),
		// Provide content service
		fx.Annotate(
			func(logger logger.Interface) content.Interface {
				return content.NewService(logger)
			},
		),
		// Provide event bus
		events.NewBus,
		// Provide debugger
		func(logger logger.Interface) debug.Debugger {
			return &debug.LogDebugger{
				Output: NewDebugLogger(logger),
			}
		},
		// Provide processors
		fx.Annotate(
			func(
				log logger.Interface,
				articleService article.Interface,
				storage storagetypes.Interface,
				sources sources.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					SourceName  string               `name:"sourceName"`
				},
			) collector.Processor {
				// Get source configuration
				source, err := sources.FindByName(params.SourceName)
				if err != nil {
					log.Error("Failed to find source", "error", err)
					return nil
				}

				log.Debug("Providing article processor")
				return &article.ArticleProcessor{
					Logger:         log,
					ArticleService: articleService,
					Storage:        storage,
					IndexName:      source.ArticleIndex,
					ArticleChan:    params.ArticleChan,
				}
			},
			fx.ResultTags(`name:"articleProcessor"`),
		),
		fx.Annotate(
			func(
				log logger.Interface,
				contentService content.Interface,
				storage storagetypes.Interface,
				sources sources.Interface,
				params struct {
					fx.In
					SourceName string `name:"sourceName"`
				},
			) collector.Processor {
				// Get source configuration
				source, err := sources.FindByName(params.SourceName)
				if err != nil {
					log.Error("Failed to find source", "error", err)
					return nil
				}

				log.Debug("Providing content processor")
				return content.NewProcessor(contentService, storage, log, source.Index)
			},
			fx.ResultTags(`name:"contentProcessor"`),
		),
		// Provide crawler with all dependencies
		fx.Annotate(
			ProvideCrawler,
			fx.ParamTags(
				``,
				``,
				``,
				``,
				`name:"articleProcessor"`,
				`name:"contentProcessor"`,
				``,
			),
		),
	),
)

// ProvideCrawler creates a new crawler instance with all dependencies.
func ProvideCrawler(
	logger logger.Interface,
	debugger debug.Debugger,
	indexManager api.IndexManager,
	sources sources.Interface,
	articleProcessor collector.Processor,
	contentProcessor collector.Processor,
	bus *events.Bus,
) Result {
	// Create crawler instance
	c := &Crawler{
		Logger:           logger,
		Debugger:         debugger,
		bus:              bus,
		indexManager:     indexManager,
		sources:          sources,
		articleProcessor: articleProcessor,
		contentProcessor: contentProcessor,
	}

	return Result{Crawler: c}
}

func NewContentProcessor(
	cfg config.Interface,
	logger logger.Interface,
	storage storagetypes.Interface,
) collector.Processor {
	service := content.NewService(logger)
	return content.NewProcessor(service, storage, logger, "content")
}

func NewArticleProcessor(
	cfg config.Interface,
	logger logger.Interface,
	storage storagetypes.Interface,
) collector.Processor {
	service := article.NewService(logger, config.DefaultArticleSelectors(), storage, "articles")
	return &article.ArticleProcessor{
		Logger:         logger,
		ArticleService: service,
		Storage:        storage,
		IndexName:      "articles",
	}
}
