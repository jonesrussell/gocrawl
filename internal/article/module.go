package article

import (
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
}

// Module provides the article module and its dependencies
var Module = fx.Module("article",
	fx.Provide(
		// Provide the article service
		func(logger logger.Interface) Interface {
			return NewService(logger)
		},
		// Provide the article processor
		func(p ProcessorParams) *Processor {
			return &Processor{
				Logger:         p.Logger,
				ArticleService: NewService(p.Logger),
				Storage:        p.Storage,
				IndexName:      p.IndexName,
				ArticleChan:    p.ArticleChan,
			}
		},
	),
)
