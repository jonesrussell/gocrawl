// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/models"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
)

// ProvideArticleService creates a new article service
func ProvideArticleService(p ContentServiceParams) ContentServiceResult {
	service := NewContentService(p.Logger, p.Storage, p.IndexName)

	return ContentServiceResult{
		Service: service,
	}
}

// ProvideArticleProcessor creates a new article processor
func ProvideArticleProcessor(p ProcessorParams) ProcessorResult {
	processor := NewProcessor(
		p.Logger,
		p.Service,
		p.Validator,
		p.Storage,
		p.IndexName,
		make(chan *models.Article, ArticleChannelBufferSize),
		p.ArticleIndexer,
		p.PageIndexer,
	)

	return ProcessorResult{
		Processor: processor,
	}
}

// Module provides the articles module's dependencies.
var Module = fx.Module("articles",
	fx.Provide(
		ProvideArticleService,
		ProvideArticleProcessor,
	),
)
