// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/processor"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the articles module's dependencies.
var Module = fx.Module("articles",
	fx.Provide(
		// Provide the article service
		fx.Annotate(
			NewContentService,
			fx.As(new(Interface)),
		),
		// Provide the article processor
		fx.Annotate(
			func(p struct {
				fx.In
				Logger         logger.Interface
				Service        Interface
				Validator      content.JobValidator
				Storage        types.Interface
				IndexName      string `name:"articleIndexName"`
				ArticleIndexer processor.Processor
				PageIndexer    processor.Processor
			}) *ArticleProcessor {
				return NewProcessor(ProcessorParams{
					Logger:         p.Logger,
					Service:        p.Service,
					Validator:      p.Validator,
					Storage:        p.Storage,
					IndexName:      p.IndexName,
					ArticleIndexer: p.ArticleIndexer,
					PageIndexer:    p.PageIndexer,
				})
			},
			fx.ResultTags(`name:"articleProcessor"`),
			fx.As(new(content.Processor)),
		),
	),
)
