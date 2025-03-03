package article

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// ProcessorParams contains the dependencies for creating a Processor
type ProcessorParams struct {
	fx.In

	Logger      logger.Interface
	Storage     storage.Interface
	IndexName   string `name:"indexName"`
	ArticleChan chan *models.Article
	Service     Interface
}

// Module provides article-related dependencies
var Module = fx.Module("article",
	fx.Provide(
		NewServiceWithConfig,
		fx.Annotate(
			NewProcessor,
			fx.As(new(models.ContentProcessor)),
		),
	),
)

// NewServiceWithConfig creates a new article service with configuration
func NewServiceWithConfig(logger logger.Interface, cfg *config.Config) Interface {
	selectors := cfg.Sources[0].Selectors.Article
	if selectors == (config.ArticleSelectors{}) {
		selectors = config.DefaultArticleSelectors()
	}
	return NewService(logger, selectors)
}

// NewProcessor creates a new article processor
func NewProcessor(p ProcessorParams) *Processor {
	return &Processor{
		Logger:         p.Logger,
		ArticleService: p.Service,
		Storage:        p.Storage,
		IndexName:      p.IndexName,
		ArticleChan:    p.ArticleChan,
	}
}
