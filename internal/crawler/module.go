// Package crawler provides the core crawling functionality for GoCrawl.
// It manages the crawling process, including URL processing, rate limiting,
// and content extraction.
package crawler

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 3
)

// Result defines the crawler module's output.
type Result struct {
	fx.Out
	Crawler Interface
}

// Module provides the crawler module for dependency injection.
var Module = fx.Module("crawler",
	fx.Provide(
		// Provide the collector
		func(logger logger.Interface, debugger debug.Debugger) *colly.Collector {
			return colly.NewCollector(
				colly.MaxDepth(DefaultMaxDepth),
				colly.Async(true),
				colly.AllowedDomains(),
				colly.ParseHTTPErrorResponse(),
			)
		},
		// Provide the crawler
		func(
			logger logger.Interface,
			indexManager api.IndexManager,
			sources sources.Interface,
			processors []common.Processor,
			bus *events.Bus,
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
			)

			logger.Info("Created crawler instance")
			return Result{
				Crawler: crawler,
			}
		},
	),
)

// ProcessorParams defines the parameters for the ProvideCrawler function.
type ProcessorParams struct {
	fx.In

	Logger           logger.Interface
	IndexManager     api.IndexManager
	Sources          sources.Interface
	ArticleProcessor common.Processor `name:"articleProcessor"`
	ContentProcessor common.Processor `name:"contentProcessor"`
	Bus              *events.Bus
}

// ProvideCrawler creates a new crawler instance.
func ProvideCrawler(params ProcessorParams) Result {
	crawler := NewCrawler(
		params.Logger,
		params.IndexManager,
		params.Sources,
		params.ArticleProcessor,
		params.ContentProcessor,
		params.Bus,
	)

	params.Logger.Info("Created crawler instance")
	return Result{
		Crawler: crawler,
	}
}

// NewCrawler creates a new crawler instance.
func NewCrawler(
	logger logger.Interface,
	indexManager api.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
) Interface {
	collector := colly.NewCollector(
		colly.MaxDepth(DefaultMaxDepth),
		colly.Async(true),
		colly.AllowedDomains(),
		colly.ParseHTTPErrorResponse(),
	)

	// Disable URL revisiting
	collector.AllowURLRevisit = false

	// Configure collector
	collector.DetectCharset = true
	collector.CheckHead = true
	collector.DisallowedDomains = []string{}

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
