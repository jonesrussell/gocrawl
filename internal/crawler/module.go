// Package crawler provides core crawling functionality.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
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

// Module provides the crawler's dependencies.
var Module = fx.Module("crawler",
	// Core dependencies
	article.Module,
	content.Module,
	fx.Provide(
		ProvideCollyDebugger,
		ProvideEventBus,
		ProvideCrawler,
		// Article channel named instance
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
		// Article index name
		fx.Annotate(
			func(cfg config.Interface) string {
				return cfg.GetElasticsearchConfig().IndexName
			},
			fx.ResultTags(`name:"indexName"`),
		),
		// Content index name
		fx.Annotate(
			func() string {
				return "content"
			},
			fx.ResultTags(`name:"contentIndex"`),
		),
		// Article processor
		fx.Annotate(
			func(
				log common.Logger,
				articleService article.Interface,
				storage types.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) collector.Processor {
				log.Debug("Providing article processor")
				return &article.ArticleProcessor{
					Logger:         log,
					ArticleService: articleService,
					Storage:        storage,
					IndexName:      params.IndexName,
					ArticleChan:    params.ArticleChan,
				}
			},
			fx.ResultTags(`group:"processors"`),
		),
		// Content processor
		fx.Annotate(
			func(
				log common.Logger,
				contentService content.Interface,
				storage types.Interface,
				params struct {
					fx.In
					IndexName string `name:"contentIndex"`
				},
			) collector.Processor {
				log.Debug("Providing content processor")
				return content.NewProcessor(contentService, storage, log, params.IndexName)
			},
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// Params defines the required dependencies for the crawler module.
type Params struct {
	fx.In

	Logger       common.Logger
	Debugger     debug.Debugger   `optional:"true"`
	IndexManager api.IndexManager `name:"indexManager"`
	Sources      sources.Interface
}

// Result contains the components provided by the crawler module.
type Result struct {
	fx.Out

	Crawler Interface
}

// ProvideCollyDebugger creates a new debugger instance.
func ProvideCollyDebugger(logger common.Logger) debug.Debugger {
	return &debug.LogDebugger{
		Output: newDebugLogger(logger),
	}
}

// ProvideEventBus creates a new event bus instance.
func ProvideEventBus() *events.Bus {
	return events.NewBus()
}

// ProvideCrawler creates a new crawler instance.
func ProvideCrawler(p Params, bus *events.Bus) (Result, error) {
	c := &Crawler{
		Logger:       p.Logger,
		Debugger:     p.Debugger,
		bus:          bus,
		indexManager: p.IndexManager,
		sources:      p.Sources,
	}
	return Result{Crawler: c}, nil
}

func NewContentProcessor(
	cfg *config.Config,
	logger logger.Interface,
) collector.Processor {
	service := content.NewService(logger)
	return content.NewProcessor(service, nil, logger, "content")
}

func NewArticleProcessor(
	cfg *config.Config,
	logger logger.Interface,
) collector.Processor {
	service := article.NewService(logger, config.DefaultArticleSelectors(), nil, "articles")
	return &article.ArticleProcessor{
		Logger:         logger,
		ArticleService: service,
		Storage:        nil,
		IndexName:      "articles",
	}
}
