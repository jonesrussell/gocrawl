// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
		// Provide the done channel
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"done"`),
		),
		// Provide the event bus
		fx.Annotate(
			func(p struct {
				fx.In
				Logger logger.Interface
			}) *events.EventBus {
				return events.NewEventBus(p.Logger)
			},
			fx.ResultTags(`name:"eventBus"`),
		),
		// Provide the index manager
		fx.Annotate(
			func(p struct {
				fx.In
				Logger logger.Interface
				Config config.Interface
			}) interfaces.IndexManager {
				client, err := storage.NewElasticsearchClient(p.Config, p.Logger)
				if err != nil {
					p.Logger.Error("Failed to create Elasticsearch client", "error", err)
					return nil
				}
				return storage.NewElasticsearchIndexManager(client, p.Logger)
			},
			fx.As(new(interfaces.IndexManager)),
		),
		// Provide the crawler
		fx.Annotate(
			func(p struct {
				fx.In
				Logger           logger.Interface
				EventBus         *events.EventBus
				IndexManager     interfaces.IndexManager
				Sources          *sources.Sources
				ArticleProcessor common.Processor
				PageProcessor    common.Processor
				Config           config.Interface
			}) crawler.Interface {
				cfg := crawlerconfig.New(
					crawlerconfig.WithMaxDepth(3),
					crawlerconfig.WithMaxConcurrency(2),
					crawlerconfig.WithRequestTimeout(30*time.Second),
					crawlerconfig.WithUserAgent("gocrawl/1.0"),
					crawlerconfig.WithRespectRobotsTxt(true),
					crawlerconfig.WithAllowedDomains([]string{"*"}),
					crawlerconfig.WithDelay(2*time.Second),
					crawlerconfig.WithRandomDelay(500*time.Millisecond),
				)
				return crawler.NewCrawler(
					p.Logger,
					p.EventBus,
					p.IndexManager,
					p.Sources,
					p.ArticleProcessor,
					p.PageProcessor,
					cfg,
				)
			},
			fx.As(new(crawler.Interface)),
		),
		// Provide the processor factory
		fx.Annotate(
			func(p struct {
				fx.In
				Logger         logger.Interface
				Config         config.Interface
				Storage        types.Interface
				ArticleService articles.Interface
				PageService    page.Interface
				IndexName      string `name:"pageIndexName"`
			}) crawler.ProcessorFactory {
				return crawler.NewProcessorFactory(crawler.ProcessorFactoryParams{
					Logger:         p.Logger,
					Config:         p.Config,
					Storage:        p.Storage,
					ArticleService: p.ArticleService,
					PageService:    p.PageService,
					IndexName:      p.IndexName,
				})
			},
			fx.As(new(crawler.ProcessorFactory)),
		),
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
		// Provide the sources
		fx.Annotate(
			func(p struct {
				fx.In
				Config config.Interface
			}) (*sources.Sources, error) {
				return sources.LoadSources(p.Config)
			},
			fx.ResultTags(`name:"sources"`),
			fx.As(new(sources.Interface)),
		),
		// Provide the job service
		fx.Annotate(
			func(p struct {
				fx.In
				Logger           logger.Interface
				Sources          *sources.Sources
				Crawler          crawler.Interface
				Done             chan struct{} `name:"done"`
				Config           config.Interface
				Storage          types.Interface
				ProcessorFactory crawler.ProcessorFactory
				SourceName       string `name:"sourceName"`
			}) (common.JobService, error) {
				if p.Sources == nil {
					return nil, fmt.Errorf("sources is nil")
				}
				return NewJobService(JobServiceParams{
					Logger:           p.Logger,
					Sources:          p.Sources,
					Crawler:          p.Crawler,
					Done:             p.Done,
					Config:           p.Config,
					Storage:          p.Storage,
					ProcessorFactory: p.ProcessorFactory,
					SourceName:       p.SourceName,
				}), nil
			},
			fx.As(new(common.JobService)),
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
			}) common.Processor {
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
		),
		// Provide the page processor
		fx.Annotate(
			func(p struct {
				fx.In
				Logger    logger.Interface
				Service   page.Interface
				Storage   types.Interface
				IndexName string `name:"pageIndexName"`
			}) common.Processor {
				return page.NewPageProcessor(page.ProcessorParams{
					Logger:    p.Logger,
					Service:   p.Service,
					Storage:   p.Storage,
					IndexName: p.IndexName,
				})
			},
			fx.ResultTags(`name:"pageProcessor"`),
		),
		// Provide the article channel
		fx.Annotate(
			func() chan *models.Article {
				return make(chan *models.Article, ArticleChannelBufferSize)
			},
			fx.ResultTags(`name:"articleChannel"`),
		),
		// Provide the processors
		fx.Annotate(
			func(p struct {
				fx.In
				ArticleProcessor common.Processor `name:"articleProcessor"`
				PageProcessor    common.Processor `name:"pageProcessor"`
			}) []common.Processor {
				return []common.Processor{
					p.ArticleProcessor,
					p.PageProcessor,
				}
			},
			fx.ResultTags(`name:"processors"`),
		),
	),
)
