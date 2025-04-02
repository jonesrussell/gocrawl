// Package crawl implements the crawl command.
package crawl

import (
	"context"

	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	mockutils "github.com/jonesrussell/gocrawl/internal/testutils"
	"go.uber.org/fx"
)

// CommandDeps holds the crawl command's dependencies
type CommandDeps struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Sources     sources.Interface
	Crawler     crawler.Interface
	Logger      types.Logger
	Config      config.Interface
	Storage     storagetypes.Interface
	Done        chan struct{}         `name:"shutdownChan"`
	Context     context.Context       `name:"crawlContext"`
	Processors  []common.Processor    `group:"processors"`
	SourceName  string                `name:"sourceName"`
	ArticleChan chan *models.Article  `name:"crawlerArticleChannel"`
	Handler     *signal.SignalHandler `name:"signalHandler"`
}

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	crawler.Module,
	sources.Module,
	config.Module,
	article.Module,
	content.Module,
	fx.Provide(
		// Command-specific dependencies
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"shutdownChan"`),
		),
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, crawler.ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
		// Provide logger
		func() common.Logger {
			return logger.NewNoOp()
		},
		// Provide debugger
		func(logger common.Logger) debug.Debugger {
			return &debug.LogDebugger{
				Output: crawler.NewDebugLogger(logger),
			}
		},
		// Provide index manager
		func() api.IndexManager {
			return &mockutils.MockIndexManager{}
		},
		// Provide sources
		func() *sources.Sources {
			return &sources.Sources{}
		},
		// Provide event bus
		func() *events.Bus {
			return events.NewBus()
		},
		// Provide article processor with correct name
		fx.Annotate(
			func(
				logger common.Logger,
				storage storagetypes.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) common.Processor {
				return &article.ArticleProcessor{
					Logger:      logger,
					Storage:     storage,
					IndexName:   params.IndexName,
					ArticleChan: params.ArticleChan,
				}
			},
			fx.ResultTags(`name:"startupArticleProcessor"`),
		),
		// Provide content processor with correct name
		fx.Annotate(
			func(
				logger common.Logger,
				storage storagetypes.Interface,
				params struct {
					fx.In
					IndexName string `name:"contentIndex"`
				},
			) common.Processor {
				return content.NewProcessor(nil, storage, logger, params.IndexName)
			},
			fx.ResultTags(`name:"startupContentProcessor"`),
		),
		// Provide index name
		fx.Annotate(
			func() string {
				return "articles"
			},
			fx.ResultTags(`name:"indexName"`),
		),
		// Provide content index name
		fx.Annotate(
			func() string {
				return "content"
			},
			fx.ResultTags(`name:"contentIndex"`),
		),
	),
)

// Params holds the crawl command's parameters
type Params struct {
	fx.In
	Sources sources.Interface `json:"sources,omitempty"`
	Logger  types.Logger

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
