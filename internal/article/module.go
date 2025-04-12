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
			ProvideArticleProcessor,
			fx.ResultTags(`group:"processors"`, `name:"articleProcessor"`),
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
		return NewService(
			logger,
			(&types.ArticleSelectors{}).Default(),
			storage,
			"articles",
		)
	}

	// Create service with default selectors
	service := NewService(
		logger,
		(&types.ArticleSelectors{}).Default(),
		storage,
		"articles",
	).(*Service)

	// Add source-specific selectors
	for _, src := range srcs {
		service.AddSourceSelectors(src.Name, src.Selectors.Article)
	}

	return service
}

// ProvideArticleProcessor creates a new article processor.
func ProvideArticleProcessor(
	logger logger.Interface,
	config config.Interface,
	storage storagetypes.Interface,
	service Interface,
) *ArticleProcessor {
	return &ArticleProcessor{
		Logger:         logger,
		ArticleService: service,
		Storage:        storage,
		IndexName:      "articles",
		ArticleChan:    make(chan *models.Article, ArticleChannelBufferSize),
		metrics:        &common.Metrics{},
	}
}
