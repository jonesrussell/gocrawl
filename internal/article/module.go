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

// ServiceParams contains the dependencies for creating a service
type ServiceParams struct {
	fx.In

	Logger logger.Interface
	Config *config.Config
	Source string `name:"sourceName"`
}

// Module provides article-related dependencies
var Module = fx.Module("article",
	fx.Provide(
		NewServiceWithConfig,
		fx.Annotate(
			NewProcessor,
			fx.As(new(models.ContentProcessor)),
			fx.ResultTags(`group:"processors"`),
		),
	),
)

// NewServiceWithConfig creates a new article service with configuration
func NewServiceWithConfig(p ServiceParams) Interface {
	// Get the source configuration
	var selectors config.ArticleSelectors
	for _, source := range p.Config.Sources {
		if source.Name == p.Source {
			selectors = source.Selectors.Article
			break
		}
	}

	if selectors == (config.ArticleSelectors{}) {
		p.Logger.Debug("Using default article selectors")
		selectors = config.DefaultArticleSelectors()
	} else {
		p.Logger.Debug("Using article selectors",
			"source", p.Source,
			"selectors", selectors)
	}

	return NewService(p.Logger, selectors)
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
