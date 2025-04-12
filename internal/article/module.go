// Package article provides functionality for processing and managing article content
// from web pages. It includes services for article extraction, processing, and storage,
// with support for configurable selectors and multiple content sources.
package article

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the size of the buffer for the article channel.
	ArticleChannelBufferSize = 100
)

// Manager handles article processing and management.
type Manager struct {
	logger  logger.Interface
	config  config.Interface
	sources sources.Interface
	storage storagetypes.Interface
}

// Module provides article-related dependencies.
var Module = fx.Options(
	fx.Provide(
		NewArticleManager,
		NewArticleService,
		fx.Annotate(
			func(
				logger logger.Interface,
				config config.Interface,
				storage storagetypes.Interface,
				jobService common.JobService,
			) (*ArticleProcessor, error) {
				selectors := types.ArticleSelectors{
					Title:         "h1",
					Description:   "meta[name=description]",
					Author:        ".author",
					PublishedTime: "time[datetime]",
					Body:          "article",
				}

				service := NewService(
					logger,
					selectors,
					storage,
					"articles",
				)

				return &ArticleProcessor{
					Logger:         logger,
					ArticleService: service,
					Storage:        storage,
					IndexName:      "articles",
					ArticleChan:    make(chan *models.Article, ArticleChannelBufferSize),
					JobService:     jobService,
					metrics:        &common.Metrics{},
				}, nil
			},
			fx.ResultTags(`name:"articleProcessor"`),
			fx.As(new(common.Processor)),
		),
	),
)

// NewArticleManager creates a new article manager.
func NewArticleManager(
	logger logger.Interface,
	config config.Interface,
	sources sources.Interface,
	storage storagetypes.Interface,
) *Manager {
	return &Manager{
		logger:  logger,
		config:  config,
		sources: sources,
		storage: storage,
	}
}

// NewArticleService creates a new article service.
func NewArticleService(
	logger logger.Interface,
	config config.Interface,
	storage storagetypes.Interface,
) Interface {
	srcs := config.GetSources()
	if len(srcs) == 0 {
		logger.Warn("No sources configured, using default selectors")
		service := NewService(
			logger,
			(&types.ArticleSelectors{}).Default(),
			storage,
			"articles",
		)
		return service
	}

	// Create service with default selectors
	service := NewService(
		logger,
		(&types.ArticleSelectors{}).Default(),
		storage,
		"articles",
	)

	// Add source-specific selectors
	for i := range srcs {
		service.AddSourceSelectors(srcs[i].Name, srcs[i].Selectors.Article)
	}

	return service
}

// NewProcessor creates a new article processor.
func NewProcessor(
	logger logger.Interface,
	service Interface,
	jobService common.JobService,
	storage storagetypes.Interface,
	indexName string,
	articleChan chan *models.Article,
) *ArticleProcessor {
	return NewArticleProcessor(ProcessorParams{
		Logger:      logger,
		Service:     service,
		JobService:  jobService,
		Storage:     storage,
		IndexName:   indexName,
		ArticleChan: articleChan,
	})
}
