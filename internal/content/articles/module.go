// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/processor"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// ProcessorParams contains dependencies for creating a processor
type ProcessorParams struct {
	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}

// ProvideArticleService creates a new article service
func ProvideArticleService(p struct {
	fx.In

	Logger    logger.Interface
	Storage   types.Interface
	IndexName string `name:"articleIndexName"`
}) (struct {
	fx.Out

	Service Interface `group:"services"`
}, error) {
	service := NewContentService(p.Logger, p.Storage, p.IndexName)

	return struct {
		fx.Out
		Service Interface `group:"services"`
	}{
		Service: service,
	}, nil
}

// ProvideArticleProcessor creates a new article processor
func ProvideArticleProcessor(p struct {
	fx.In

	Logger         logger.Interface
	Service        Interface
	Validator      content.JobValidator
	Storage        types.Interface
	IndexName      string `name:"articleIndexName"`
	ArticleIndexer processor.Processor
	PageIndexer    processor.Processor
}) (struct {
	fx.Out

	Processor content.Processor `name:"articleProcessor"`
}, error) {
	processor := NewProcessor(
		p.Logger,
		p.Service,
		p.Validator,
		p.Storage,
		p.IndexName,
		make(chan *models.Article, 100),
		p.ArticleIndexer,
		p.PageIndexer,
	)

	return struct {
		fx.Out
		Processor content.Processor `name:"articleProcessor"`
	}{
		Processor: processor,
	}, nil
}

// Module provides the articles module's dependencies.
var Module = fx.Module("articles",
	fx.Provide(
		ProvideArticleService,
		ProvideArticleProcessor,
	),
)
