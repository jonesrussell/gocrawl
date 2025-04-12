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
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
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
	// DefaultInitTimeout is the default timeout for module initialization.
	DefaultInitTimeout = 30 * time.Second
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

		// Article channel with lifecycle management
		fx.Annotate(
			func(lc fx.Lifecycle) chan *models.Article {
				ch := make(chan *models.Article, ArticleChannelBufferSize)

				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						close(ch)
						return nil
					},
				})

				return ch
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),

		// Sources with error handling
		fx.Annotate(
			func(config config.Interface, logger logger.Interface, sourceName string) (*sources.Sources, error) {
				logger.Debug("Loading source from configuration", "source", sourceName)

				src, err := sources.LoadSources(config)
				if err != nil {
					logger.Error("Failed to load sources",
						"error", err,
					)
					return nil, fmt.Errorf("failed to load sources: %w", err)
				}

				// Get all sources to find the requested one
				sourceConfigs, err := src.GetSources()
				if err != nil {
					logger.Error("Failed to get source configurations",
						"error", err,
					)
					return nil, fmt.Errorf("failed to get source configurations: %w", err)
				}

				// Find the requested source
				var requestedSource sources.Config
				for _, cfg := range sourceConfigs {
					if cfg.Name == sourceName {
						requestedSource = cfg
						break
					}
				}

				if requestedSource.Name == "" {
					logger.Error("Source not found",
						"source", sourceName,
						"available_sources", sourceConfigs,
					)
					return nil, fmt.Errorf("source not found: %s", sourceName)
				}

				// Log source details
				logger.Info("Loaded source configuration",
					"name", requestedSource.Name,
					"url", requestedSource.URL,
					"index", requestedSource.Index,
					"max_depth", requestedSource.MaxDepth,
					"rate_limit", requestedSource.RateLimit,
				)

				// Create a new Sources instance with just the requested source
				return sources.NewSources(&requestedSource, logger), nil
			},
			fx.ParamTags(``, ``, `name:"sourceName"`),
		),

		// Event bus with error handling
		func(logger logger.Interface) (*events.EventBus, error) {
			logger.Debug("Creating event bus")
			bus := events.NewEventBus(logger)
			if bus == nil {
				logger.Error("Failed to create event bus")
				return nil, errors.New("failed to create event bus")
			}
			return bus, nil
		},

		// Index manager with error handling
		fx.Annotate(
			func(
				config config.Interface,
				logger logger.Interface,
				client *elasticsearch.Client,
			) (interfaces.IndexManager, error) {
				logger.Debug("Creating Elasticsearch index manager")

				if client == nil {
					logger.Error("Elasticsearch client not initialized")
					return nil, errors.New("elasticsearch client not initialized")
				}

				manager := storage.NewElasticsearchIndexManager(client, logger)
				if manager == nil {
					logger.Error("Failed to create Elasticsearch index manager")
					return nil, errors.New("failed to create Elasticsearch index manager")
				}

				return manager, nil
			},
			fx.ResultTags(`name:"indexManager"`),
			fx.As(new(interfaces.IndexManager)),
		),

		// Signal handler with lifecycle integration
		fx.Annotate(
			func(lc fx.Lifecycle, logger logger.Interface) signal.Interface {
				handler := signal.NewSignalHandler(logger)

				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						handler.RequestShutdown()
						return handler.Wait()
					},
				})

				return handler
			},
			fx.As(new(signal.Interface)),
		),

		// Article processor
		article.ProvideArticleProcessor,

		// Content processor with configuration
		func(
			logger logger.Interface,
			service content.Interface,
			storage storagetypes.Interface,
			sources *sources.Sources,
		) (*content.ContentProcessor, error) {
			if sources == nil {
				return nil, errors.New("sources not initialized")
			}

			sourceConfigs, err := sources.GetSources()
			if err != nil {
				return nil, fmt.Errorf("failed to get sources: %w", err)
			}

			if len(sourceConfigs) == 0 {
				return nil, errors.New("no sources configured")
			}

			indexName := sourceConfigs[0].Index
			if indexName == "" {
				indexName = "content" // Fallback to default
			}

			return content.NewContentProcessor(content.ProcessorParams{
				Logger:    logger,
				Service:   service,
				Storage:   storage,
				IndexName: indexName,
			}), nil
		},

		// Processors slice with error handling
		func(
			articleProcessor *article.ArticleProcessor,
			contentProcessor *content.ContentProcessor,
			logger logger.Interface,
		) ([]common.Processor, error) {
			if articleProcessor == nil || contentProcessor == nil {
				return nil, errors.New("processors not initialized")
			}

			return []common.Processor{articleProcessor, contentProcessor}, nil
		},
	),

	// Include crawler module after processors are provided
	crawler.Module,

	// Add initialization timeout
	fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				initCtx, cancel := context.WithTimeout(ctx, DefaultInitTimeout)
				defer cancel()

				select {
				case <-initCtx.Done():
					return fmt.Errorf("module initialization timed out after %v", DefaultInitTimeout)
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		})
	}),
)
