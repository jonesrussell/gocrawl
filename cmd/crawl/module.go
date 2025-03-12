// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the crawl command dependencies
var Module = fx.Module("crawl",
	// Core dependencies
	config.Module,
	logger.Module,
	storage.Module,
	sources.Module,
	api.Module,

	// Feature modules
	article.Module,
	content.Module,
	collector.Module(),
	crawler.Module,

	fx.Provide(
		fx.Annotate(
			func() chan struct{} { return make(chan struct{}) },
			fx.ResultTags(`name:"crawlDone"`),
		),
		fx.Annotate(
			func() string { return sourceName },
			fx.ResultTags(`name:"sourceName"`),
		),
		fx.Annotate(
			func(sources sources.Interface) (string, string) {
				src, srcErr := sources.FindByName(sourceName)
				if srcErr != nil {
					return "", ""
				}
				return src.Index, src.ArticleIndex
			},
			fx.ParamTags(`name:"sourceManager"`),
			fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
		),
		func() chan *models.Article {
			return make(chan *models.Article, app.DefaultChannelBufferSize)
		},
	),
)
