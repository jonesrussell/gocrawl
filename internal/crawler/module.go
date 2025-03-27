// Package crawler provides core crawling functionality.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

const (
	// articleChannelBufferSize is the buffer size for the article channel.
	articleChannelBufferSize = 100
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
	GetMetrics() *Metrics
}

// Module provides the crawler's dependencies.
var Module = fx.Module("crawler",
	// Core dependencies
	common.Module,
	article.Module,
	content.Module,
	fx.Provide(
		provideCollectorConfig,
		provideCollyDebugger,
		provideEventBus,
		provideCrawler,
		// Content service
		fx.Annotate(
			content.NewService,
			fx.As(new(content.Interface)),
		),
		// Core dependencies
		fx.Annotate(
			func(log common.Logger) chan struct{} {
				log.Debug("Providing Done channel")
				return make(chan struct{})
			},
			fx.ResultTags(`name:"crawlDone"`),
		),
		// Article channel named instance
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, articleChannelBufferSize)
			},
			fx.ResultTags(`name:"articleChannel"`),
		),
		// Article index name
		fx.Annotate(
			func() string {
				return "articles"
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
					ArticleChan chan *models.Article `name:"articleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) models.ContentProcessor {
				log.Debug("Providing article processor")
				return &article.Processor{
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
			) models.ContentProcessor {
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
	Debugger     debug.Debugger    `optional:"true"`
	IndexManager api.IndexManager  `name:"indexManager"`
	Sources      sources.Interface `name:"testSourceManager"`
}

// Result contains the components provided by the crawler module.
type Result struct {
	fx.Out

	Crawler Interface
}

// CrawlDeps defines the dependencies required for crawling.
type CrawlDeps struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Sources     sources.Interface `name:"testSourceManager"`
	Crawler     Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     types.Interface
	Done        chan struct{}             `name:"crawlDone"`
	Context     context.Context           `name:"crawlContext"`
	Processors  []models.ContentProcessor `group:"processors"`
	SourceName  string                    `name:"sourceName"`
	ArticleChan chan *models.Article      `name:"articleChannel"`
	Handler     *signal.SignalHandler     `name:"signalHandler"`
}

// provideCollectorConfig creates a new collector configuration.
func provideCollectorConfig(cfg config.Interface, logger common.Logger) *collector.Config {
	crawlerCfg := cfg.GetCrawlerConfig()
	return &collector.Config{
		BaseURL:     crawlerCfg.BaseURL,
		MaxDepth:    crawlerCfg.MaxDepth,
		RateLimit:   crawlerCfg.RateLimit.String(),
		Parallelism: 1,
		RandomDelay: crawlerCfg.RandomDelay,
		Logger:      logger,
		Source: config.Source{
			URL:       crawlerCfg.BaseURL,
			MaxDepth:  crawlerCfg.MaxDepth,
			RateLimit: crawlerCfg.RateLimit,
		},
	}
}

// provideEventBus creates a new event bus instance.
func provideEventBus() *events.Bus {
	return events.NewBus()
}

// provideCollyDebugger creates a new debugger instance.
func provideCollyDebugger(logger common.Logger) debug.Debugger {
	return &debug.LogDebugger{
		Output: newDebugLogger(logger),
	}
}

// provideCrawler creates a new crawler instance.
func provideCrawler(p Params, bus *events.Bus) (Result, error) {
	c := &Crawler{
		Logger:       p.Logger,
		Debugger:     p.Debugger,
		bus:          bus,
		indexManager: p.IndexManager,
		sources:      p.Sources,
	}
	return Result{Crawler: c}, nil
}
