// Package processor provides functionality for processing web content.
// It handles both article and general content processing, with support
// for configurable selectors and multiple content types.
package processor

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the processor module's dependencies.
var Module = fx.Module("processor",
	fx.Provide(
		// Provide article processor
		fx.Annotate(
			func(
				logger common.Logger,
				storage types.Interface,
				params struct {
					fx.In
					ArticleChan chan *models.Article `name:"crawlerArticleChannel"`
					IndexName   string               `name:"indexName"`
				},
			) common.Processor {
				// Create article service
				articleService := article.NewService(logger, config.DefaultArticleSelectors(), storage, params.IndexName)
				logger.Debug("Created article service", "type", fmt.Sprintf("%T", articleService))

				// Create article processor
				articleProcessor := &article.ArticleProcessor{
					Logger:         logger,
					ArticleService: articleService,
					Storage:        storage,
					IndexName:      params.IndexName,
					ArticleChan:    params.ArticleChan,
				}
				logger.Debug("Created article processor", "type", fmt.Sprintf("%T", articleProcessor))
				return articleProcessor
			},
			fx.ResultTags(`name:"articleProcessor"`),
		),
		// Provide content processor
		fx.Annotate(
			func(
				logger common.Logger,
				storage types.Interface,
				params struct {
					fx.In
					IndexName string `name:"indexName"`
				},
			) common.Processor {
				// Create content service
				contentService := content.NewService(logger)
				logger.Debug("Created content service", "type", fmt.Sprintf("%T", contentService))

				// Create content processor
				contentProcessor := &ContentProcessor{
					Logger:         logger,
					Storage:        storage,
					IndexName:      params.IndexName,
					ContentService: contentService,
					metrics:        &common.Metrics{},
				}
				logger.Debug("Created content processor", "type", fmt.Sprintf("%T", contentProcessor))
				return contentProcessor
			},
			fx.ResultTags(`name:"contentProcessor"`),
		),
	),
)
