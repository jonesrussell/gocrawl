// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultInitTimeout is the default timeout for module initialization.
	DefaultInitTimeout = 30 * time.Second
)

// Module provides the crawl command module for dependency injection
var Module = fx.Options(
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	article.Module,
	content.Module,

	// Provide the context
	fx.Provide(context.Background),

	// Provide the article channel
	fx.Provide(func() chan *models.Article {
		return make(chan *models.Article, ArticleChannelBufferSize)
	}),

	// Provide the page index name
	fx.Provide(fx.Annotate(
		func() string {
			return "page"
		},
		fx.ResultTags(`name:"pageIndexName"`),
	)),

	// Provide the sources
	fx.Provide(func(logger logger.Interface, cfg config.Interface) (*sources.Sources, error) {
		return sources.LoadSources(cfg)
	}),

	// Provide the event bus
	fx.Provide(events.NewEventBus),

	// Provide the signal handler
	fx.Provide(func() chan struct{} {
		return make(chan struct{})
	}),

	// Provide the processors
	fx.Provide(func(
		logger logger.Interface,
		articleChannel chan *models.Article,
	) []common.Processor {
		return []common.Processor{
			&common.NoopProcessor{}, // Article processor
			&common.NoopProcessor{}, // Page processor
		}
	}),

	// Provide the processor factory
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			config config.Interface,
			storage types.Interface,
			articleService article.Interface,
			pageService content.Interface,
			indexName string,
			articleChannel chan *models.Article,
		) crawler.ProcessorFactory {
			return crawler.NewProcessorFactory(crawler.ProcessorFactoryParams{
				Logger:         logger,
				Config:         config,
				Storage:        storage,
				ArticleService: articleService,
				PageService:    pageService,
				IndexName:      indexName,
				ArticleChannel: articleChannel,
			})
		},
		fx.ParamTags(``, ``, ``, ``, ``, `name:"pageIndexName"`, ``),
	)),

	// Provide the job service
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			storage types.Interface,
			sources *sources.Sources,
			crawler crawler.Interface,
			done chan struct{},
			config config.Interface,
			processorFactory crawler.ProcessorFactory,
			sourceName string,
		) common.JobService {
			return NewJobService(JobServiceParams{
				Logger:           logger,
				Sources:          sources,
				Crawler:          crawler,
				Done:             done,
				Config:           config,
				Storage:          storage,
				ProcessorFactory: processorFactory,
				SourceName:       sourceName,
			})
		},
		fx.As(new(common.JobService)),
		fx.ParamTags(``, ``, ``, ``, ``, ``, ``, `name:"sourceName"`),
	)),

	// Provide the crawler
	fx.Provide(func(
		ctx context.Context,
		logger logger.Interface,
		sources *sources.Sources,
		eventBus *events.EventBus,
		processors []common.Processor,
		cfg *crawlerconfig.Config,
		storage types.Interface,
	) (crawler.Interface, error) {
		return SetupCollector(
			ctx,
			logger,
			storage.GetIndexManager(),
			sources,
			eventBus,
			processors[0],
			processors[1],
			cfg,
		)
	}),
)
