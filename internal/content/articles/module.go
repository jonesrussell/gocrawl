// Package articles provides functionality for processing and managing article content.
package articles

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
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
				JobService     common.JobService
				Storage        types.Interface
				IndexName      string `name:"articleIndexName"`
				ArticleChannel chan *models.Article
			}) *ArticleProcessor {
				return NewProcessor(ProcessorParams{
					Logger:         p.Logger,
					Service:        p.Service,
					JobService:     p.JobService,
					Storage:        p.Storage,
					IndexName:      p.IndexName,
					ArticleChannel: p.ArticleChannel,
				})
			},
			fx.ResultTags(`name:"articleProcessor"`),
			fx.As(new(common.Processor)),
		),
	),
)
