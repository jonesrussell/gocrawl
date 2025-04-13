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
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
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
	// Provide the article channel
	fx.Provide(func() chan *models.Article {
		return make(chan *models.Article, 100)
	}),

	// Provide the content index name
	fx.Provide(fx.Annotate(
		func() string {
			return "content"
		},
		fx.ResultTags(`name:"contentIndexName"`),
	)),

	// Provide the sources
	fx.Provide(func(logger logger.Interface) *sources.Sources {
		return sources.NewSources(&sources.Config{}, logger)
	}),

	// Provide the event bus
	fx.Provide(func(logger logger.Interface) *events.EventBus {
		return events.NewEventBus(logger)
	}),

	// Provide the signal handler
	fx.Provide(func() chan struct{} {
		return make(chan struct{})
	}),

	// Provide the processors
	fx.Provide(func(
		logger logger.Interface,
		articleChannel chan *models.Article,
		indexManager interfaces.IndexManager,
	) []common.Processor {
		return []common.Processor{
			&common.NoopProcessor{}, // Article processor
			&common.NoopProcessor{}, // Content processor
		}
	}),

	// Provide the processor factory
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			config config.Interface,
			storage types.Interface,
			articleService article.Interface,
			contentService content.Interface,
			indexName string,
			articleChannel chan *models.Article,
		) crawler.ProcessorFactory {
			return crawler.NewProcessorFactory(crawler.ProcessorFactoryParams{
				Logger:         logger,
				Config:         config,
				Storage:        storage,
				ArticleService: articleService,
				ContentService: contentService,
				IndexName:      indexName,
				ArticleChannel: articleChannel,
			})
		},
		fx.ParamTags(``, ``, ``, ``, ``, `name:"contentIndexName"`, ``),
	)),

	// Provide the job service
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			storage types.Interface,
			indexManager interfaces.IndexManager,
			sources *sources.Sources,
			crawler crawler.Interface,
			done chan struct{},
			config config.Interface,
			processorFactory crawler.ProcessorFactory,
		) common.JobService {
			return NewJobService(JobServiceParams{
				Logger:           logger,
				Sources:          sources,
				Crawler:          crawler,
				Done:             done,
				Config:           config,
				Storage:          storage,
				ProcessorFactory: processorFactory,
			})
		},
		fx.As(new(common.JobService)),
	)),

	// Provide the crawler
	fx.Provide(func(
		ctx context.Context,
		logger logger.Interface,
		indexManager interfaces.IndexManager,
		sources *sources.Sources,
		eventBus *events.EventBus,
		processors []common.Processor,
		cfg *crawlerconfig.Config,
	) (crawler.Interface, error) {
		return SetupCollector(
			ctx,
			logger,
			indexManager,
			sources,
			eventBus,
			processors[0],
			processors[1],
			cfg,
		)
	}),
)

// ProcessorParams holds parameters for creating processors.
type ProcessorParams struct {
	fx.In
	Logger         logger.Interface
	Config         config.Interface
	Storage        types.Interface
	ArticleService article.Interface
	ContentService content.Interface
	IndexName      string `name:"contentIndexName"`
	ArticleChannel chan *models.Article
}
