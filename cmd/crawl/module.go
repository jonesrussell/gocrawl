// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"

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

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
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
		func(config config.Interface) logger.Params {
			level := logger.InfoLevel
			if config.GetAppConfig().Debug {
				level = logger.DebugLevel
			}
			return logger.Params{
				Config: &logger.Config{
					Level:            level,
					Development:      true,
					Encoding:         "console",
					OutputPaths:      []string{"stdout"},
					ErrorOutputPaths: []string{"stderr"},
					EnableColor:      true,
				},
			}
		},

		// Article channel
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, ArticleChannelBufferSize)
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
		fx.Annotate(
			signal.NewSignalHandler,
			fx.As(new(signal.Interface)),
		),

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
)
