// Package crawl implements the crawl command.
package crawl

import (
	"context"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
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
	Sources     sources.Interface `name:"testSourceManager"`
	Crawler     crawler.Interface
	Logger      common.Logger
	Config      config.Interface
	Storage     types.Interface
	Done        chan struct{}             `name:"commandDone"`
	Context     context.Context           `name:"crawlContext"`
	Processors  []models.ContentProcessor `group:"processors"`
	SourceName  string                    `name:"sourceName"`
	ArticleChan chan *models.Article      `name:"commandArticleChannel"`
	Handler     *signal.SignalHandler     `name:"signalHandler"`
}

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	crawler.Module,
	fx.Provide(
		// Command-specific dependencies
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"commandDone"`),
		),
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, crawler.ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"commandArticleChannel"`),
		),
	),
)
