// Package crawler provides the core crawling functionality for GoCrawl.
// It manages the crawling process, including URL processing, rate limiting,
// and content extraction.
package crawler

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
)

// Result defines the crawler module's output.
type Result struct {
	fx.Out
	Crawler Interface
}

// ProvideCrawler creates a new crawler instance with the given dependencies.
func ProvideCrawler(
	logger logger.Interface,
	indexManager storagetypes.IndexManager,
	sources sources.Interface,
	processors []common.Processor,
	bus *events.Bus,
	cfg *crawler.Config,
) Result {
	// Find article and content processors
	var articleProcessor, contentProcessor common.Processor
	for _, p := range processors {
		if p.ContentType() == common.ContentTypeArticle {
			articleProcessor = p
		} else if p.ContentType() == common.ContentTypePage {
			contentProcessor = p
		}
	}

	crawler := NewCrawler(
		logger,
		indexManager,
		sources,
		articleProcessor,
		contentProcessor,
		bus,
		cfg,
	)

	logger.Info("Created crawler instance")
	return Result{
		Crawler: crawler,
	}
}

// Module provides the crawler module for dependency injection.
var Module = fx.Module("crawler",
	fx.Provide(
		// Provide the collector
		func(logger logger.Interface, debugger debug.Debugger, cfg *crawler.Config) *colly.Collector {
			return colly.NewCollector(
				colly.MaxDepth(cfg.MaxDepth),
				colly.Async(true),
				colly.AllowedDomains(cfg.AllowedDomains...),
				colly.ParseHTTPErrorResponse(),
			)
		},
		// Provide the crawler
		ProvideCrawler,
	),
)

// NewCrawler creates a new crawler instance.
func NewCrawler(
	logger logger.Interface,
	indexManager storagetypes.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
	cfg *crawler.Config,
) Interface {
	collector := colly.NewCollector(
		colly.MaxDepth(cfg.MaxDepth),
		colly.Async(true),
		colly.AllowedDomains(cfg.AllowedDomains...),
		colly.ParseHTTPErrorResponse(),
	)

	// Disable URL revisiting
	collector.AllowURLRevisit = false

	// Configure collector
	collector.DetectCharset = true
	collector.CheckHead = true
	collector.DisallowedDomains = cfg.DisallowedDomains
	collector.UserAgent = cfg.UserAgent
	collector.IgnoreRobotsTxt = !cfg.RespectRobotsTxt

	// Set rate limiting
	if cfg.Delay > 0 {
		collector.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			RandomDelay: cfg.RandomDelay,
			Parallelism: cfg.MaxConcurrency,
		})
	}

	crawler := &Crawler{
		Logger:           logger,
		indexManager:     indexManager,
		sources:          sources,
		articleProcessor: articleProcessor,
		contentProcessor: contentProcessor,
		bus:              bus,
		collector:        collector,
	}

	// Set up callbacks
	collector.OnRequest(func(r *colly.Request) {
		crawler.Logger.Info("Visiting", "url", r.URL.String())
	})

	collector.OnResponse(func(r *colly.Response) {
		crawler.Logger.Info("Visited", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	collector.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", err)
	})

	return crawler
}
