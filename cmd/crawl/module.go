// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
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
	fx.Provide(func() string {
		return "content"
	}),

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

	// Provide the job service
	fx.Provide(func(
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
	}),

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

// provideProcessors creates and provides the content processors.
func provideProcessors(p ProcessorParams) ([]common.Processor, error) {
	articleProcessor := article.NewProcessor(
		p.Logger,
		p.ArticleService,
		nil, // JobService will be injected later
		p.Storage,
		p.IndexName,
		p.ArticleChannel,
	)

	contentProcessor := content.NewContentProcessor(content.ProcessorParams{
		Logger:    p.Logger,
		Service:   p.ContentService,
		Storage:   p.Storage,
		IndexName: p.IndexName,
	})

	return []common.Processor{
		articleProcessor,
		contentProcessor,
	}, nil
}

// provideArticleChannel provides the article channel for communication between components.
func provideArticleChannel() chan *models.Article {
	return make(chan *models.Article, ArticleChannelBufferSize)
}

// provideContentIndexName provides the content index name.
func provideContentIndexName() string {
	return "content"
}

// provideEventBus creates and provides the event bus.
func provideEventBus(logger logger.Interface) (*events.EventBus, error) {
	bus := events.NewEventBus(logger)
	if bus == nil {
		return nil, errors.New("failed to create event bus")
	}
	return bus, nil
}

// provideSignalHandler creates and provides the signal handler.
func provideSignalHandler(logger logger.Interface) signal.Interface {
	return signal.NewSignalHandler(logger)
}

// provideDoneChannel provides the done channel for graceful shutdown.
func provideDoneChannel() chan struct{} {
	return make(chan struct{})
}

// provideSourceConfig loads and provides the source configuration.
func provideSourceConfig(config config.Interface, logger logger.Interface, sourceName string) (*sources.Sources, error) {
	src, err := sources.LoadSources(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	sourceConfigs, err := src.GetSources()
	if err != nil {
		return nil, fmt.Errorf("failed to get source configurations: %w", err)
	}

	var requestedSource sources.Config
	for _, cfg := range sourceConfigs {
		if cfg.Name == sourceName {
			requestedSource = cfg
			break
		}
	}

	if requestedSource.Name == "" {
		return nil, fmt.Errorf("source not found: %s", sourceName)
	}

	return sources.NewSources(&requestedSource, logger), nil
}

// provideJobService creates and provides the job service.
func provideJobService(p struct {
	fx.In
	Logger           logger.Interface
	Sources          *sources.Sources
	Crawler          crawler.Interface
	Done             chan struct{}
	Config           config.Interface
	Storage          types.Interface
	ProcessorFactory crawler.ProcessorFactory
}) (common.JobService, error) {
	// Get the underlying Elasticsearch client from the storage interface
	client, ok := p.Storage.(*elasticsearch.Client)
	if !ok {
		return nil, fmt.Errorf("storage interface must be an Elasticsearch client")
	}

	return NewJobService(JobServiceParams{
		Logger:           p.Logger,
		Sources:          p.Sources,
		Crawler:          p.Crawler,
		Done:             p.Done,
		Config:           p.Config,
		Client:           client,
		ProcessorFactory: p.ProcessorFactory,
	}), nil
}
