// Package crawl implements the crawl command.
package crawl

import (
	"context"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// CommandDeps holds the crawl command's dependencies
type CommandDeps struct {
	fx.In

	Lifecycle   fx.Lifecycle
	Sources     sources.Interface
	Crawler     crawler.Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     types.Interface
	Done        chan struct{}
	Context     context.Context
	Processors  []collector.Processor `group:"processors"`
	SourceName  string
	ArticleChan chan *models.Article
	Handler     *signal.SignalHandler
}

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	crawler.Module,
	fx.Provide(
		// Command-specific dependencies
		func() chan struct{} {
			return make(chan struct{})
		},
		func() chan *models.Article {
			return make(chan *models.Article, crawler.ArticleChannelBufferSize)
		},
	),
)

type Params struct {
	fx.In
	Sources sources.Interface `json:"sources,omitempty"`
	Logger  common.Logger

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle `json:"lifecycle,omitempty"`

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface `json:"crawler_instance,omitempty"`

	// Config holds the application configuration
	Config config.Interface `json:"config,omitempty"`

	// Context provides the context for the crawl operation
	Context context.Context `name:"crawlContext" json:"context,omitempty"`

	// Processors is a slice of content processors, injected as a group
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone" json:"done,omitempty"`
}
