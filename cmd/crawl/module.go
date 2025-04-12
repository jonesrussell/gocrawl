// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"time"

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
var Module = fx.Module("crawl",
	// Include required modules
	config.Module,
	storage.Module,
	logger.Module,
	sources.Module,
	article.Module,
	content.Module,
	crawler.Module,

	fx.Provide(
		// Provide article channel
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, 100)
			},
			fx.ResultTags(`name:"crawlerArticleChannel"`),
		),
		// Provide sources
		fx.Annotate(
			func(config config.Interface, logger logger.Interface, sourceName string) (*sources.Sources, error) {
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
			},
			fx.ParamTags(``, ``, `name:"sourceName"`),
		),
		// Provide event bus
		func(logger logger.Interface) (*events.EventBus, error) {
			bus := events.NewEventBus(logger)
			if bus == nil {
				return nil, errors.New("failed to create event bus")
			}
			return bus, nil
		},
		// Provide signal handler
		fx.Annotate(
			func(logger logger.Interface) signal.Interface {
				return signal.NewSignalHandler(logger)
			},
			fx.As(new(signal.Interface)),
		),
		// Provide processors
		fx.Annotate(
			func(articleProcessor, contentProcessor common.Processor) []common.Processor {
				if articleProcessor == nil || contentProcessor == nil {
					return nil
				}
				return []common.Processor{articleProcessor, contentProcessor}
			},
			fx.ParamTags(`name:"articleProcessor" group:"processors"`, `name:"contentProcessor" group:"processors"`),
		),
	),
)
