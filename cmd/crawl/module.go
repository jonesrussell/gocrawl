// Package crawl implements the crawl command.
package crawl

import (
	"context"

	"github.com/gocolly/colly/v2/debug"
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
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
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
	Done        chan struct{}        `name:"shutdownChan"`
	Context     context.Context      `name:"crawlContext"`
	Processors  []common.Processor   `group:"processors"`
	SourceName  string               `name:"sourceName"`
	ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
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

	// Provide debugger
	fx.Provide(
		func() debug.Debugger {
			return &debug.LogDebugger{}
		},
	),

	// Provide event bus
	fx.Provide(
		fx.Annotate(
			events.NewBus,
			fx.ResultTags(`name:"eventBus"`),
		),
	),

	// Provide processors
	fx.Provide(
		fx.Annotate(
			func(
				logger common.Logger,
				storage storagetypes.Interface,
				contentService content.Interface,
				params struct {
					fx.In
					IndexName string `name:"contentIndex"`
				},
			) common.Processor {
				processor := &content.ContentProcessor{
					Logger:         logger,
					ContentService: contentService,
					Storage:        storage,
					IndexName:      params.IndexName,
				}
				if processor == nil {
					panic("failed to create content processor")
				}
				return processor
			},
			fx.ResultTags(`name:"contentProcessor"`),
		),
	),

	// Provide index names
	fx.Provide(
		fx.Annotate(
			func() string {
				return "articles"
			},
			fx.ResultTags(`name:"indexName"`),
		),
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
