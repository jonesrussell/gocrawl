// Package crawl implements the crawl command.
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

// CommandDeps holds the dependencies for the crawl command.
type CommandDeps struct {
	fx.In

	Context     context.Context `name:"crawlContext"`
	SourceName  string          `name:"sourceName"`
	Config      config.Interface
	Logger      logger.Interface
	Storage     storagetypes.Interface
	Crawler     crawler.Interface
	Sources     sources.Interface
	Handler     *signal.SignalHandler `name:"signalHandler"`
	ArticleChan chan *models.Article  `name:"crawlerArticleChannel"`
	Processors  []common.Processor    `group:"processors"`
}

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	// Core dependencies
	config.Module,
	logger.Module,
	sources.Module,
	storage.Module,
	crawler.Module,
	article.Module,
	content.Module,

	// Provide command channels
	fx.Provide(
		NewCrawler,
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"shutdownChan"`),
		),
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
	),

	// Provide index names
	fx.Provide(
		fx.Annotate(
			func(sources sources.Interface, sourceName string) string {
				source := sources.FindByName(sourceName)
				if source == nil {
					panic(fmt.Sprintf("failed to get source %s: source not found", sourceName))
				}
				return source.Index
			},
			fx.ParamTags(`name:"sourceName"`),
			fx.ResultTags(`name:"indexName"`),
		),
		fx.Annotate(
			func(sources sources.Interface, sourceName string) string {
				source := sources.FindByName(sourceName)
				if source == nil {
					panic(fmt.Sprintf("failed to get source %s: source not found", sourceName))
				}
				return source.Index
			},
			fx.ParamTags(`name:"sourceName"`),
			fx.ResultTags(`name:"contentIndex"`),
		),
	),

	// Provide crawler instance
	fx.Provide(
		fx.Annotate(
			func(
				logger logger.Interface,
				indexManager storagetypes.IndexManager,
				sources sources.Interface,
				processors []common.Processor,
				bus *events.Bus,
			) crawler.Interface {
				// Find article and content processors
				var articleProcessor, contentProcessor common.Processor
				for _, p := range processors {
					if p.ContentType() == common.ContentTypeArticle {
						articleProcessor = p
					} else if p.ContentType() == common.ContentTypePage {
						contentProcessor = p
					}
				}

				return crawler.NewCrawler(
					logger,
					indexManager,
					sources,
					articleProcessor,
					contentProcessor,
					bus,
				)
			},
			fx.ResultTags(`name:"crawler"`),
		),
	),
)

// Params holds the crawl command's parameters
type Params struct {
	fx.In
	Sources sources.Interface `json:"sources,omitempty"`
	Logger  logger.Interface

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle `json:"lifecycle,omitempty"`

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface `json:"crawler_instance,omitempty"`

	// Config holds the application configuration
	Config config.Interface `json:"config,omitempty"`

	// Context provides the context for the crawl operation
	Context context.Context `name:"crawlContext" json:"context,omitempty"`

	// Processors is a slice of content processors, injected as a group
	Processors []common.Processor `group:"processors" json:"processors,omitempty"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone" json:"done,omitempty"`
}
