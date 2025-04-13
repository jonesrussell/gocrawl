// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultInitTimeout is the default timeout for module initialization.
	DefaultInitTimeout = 30 * time.Second
)

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	fx.Provide(
		// Provide the article service
		fx.Annotate(
			func(p struct {
				fx.In
				Logger    logger.Interface
				Storage   types.Interface
				IndexName string `name:"articleIndexName"`
			}) articles.Interface {
				return articles.NewContentService(articles.ServiceParams{
					Logger:    p.Logger,
					Storage:   p.Storage,
					IndexName: p.IndexName,
				})
			},
			fx.ResultTags(`name:"articleService"`),
		),
		// Provide the page service
		fx.Annotate(
			func(p struct {
				fx.In
				Logger    logger.Interface
				Storage   types.Interface
				IndexName string `name:"pageIndexName"`
			}) page.Interface {
				return page.NewContentService(page.ServiceParams{
					Logger:    p.Logger,
					Storage:   p.Storage,
					IndexName: p.IndexName,
				})
			},
			fx.ResultTags(`name:"pageService"`),
		),
		// Provide the article processor
		fx.Annotate(
			func(p struct {
				fx.In
				Logger         logger.Interface
				Service        articles.Interface
				JobService     common.JobService
				Storage        types.Interface
				IndexName      string `name:"articleIndexName"`
				ArticleChannel chan *models.Article
			}) *articles.ArticleProcessor {
				return articles.NewProcessor(articles.ProcessorParams{
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
		// Provide the page processor
		fx.Annotate(
			func(p struct {
				fx.In
				Logger    logger.Interface
				Service   page.Interface
				Storage   types.Interface
				IndexName string `name:"pageIndexName"`
			}) *page.PageProcessor {
				return page.NewPageProcessor(page.ProcessorParams{
					Logger:    p.Logger,
					Service:   p.Service,
					Storage:   p.Storage,
					IndexName: p.IndexName,
				})
			},
			fx.ResultTags(`name:"pageProcessor"`),
			fx.As(new(common.Processor)),
		),
	),
)
