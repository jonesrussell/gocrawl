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
	"github.com/jonesrussell/gocrawl/internal/sources"
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

// Module provides the crawler module and its dependencies.
var Module = fx.Module("crawler",
	fx.Provide(
		// Provide the crawler implementation
		fx.Annotate(
			ProvideCrawler,
			fx.ParamTags(
				``,
				``,
				``,
				``,
				`name:"startupArticleProcessor"`,
				`name:"startupContentProcessor"`,
				`name:"eventBus"`,
			),
		),
	),
)

// ProvideCrawler creates a new crawler instance.
func ProvideCrawler(
	logger common.Logger,
	debugger debug.Debugger,
	indexManager api.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
) Result {
	crawler := NewCrawler(
		logger,
		debugger,
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
}

// NewCrawler creates a new crawler instance.
func NewCrawler(
	logger common.Logger,
	debugger debug.Debugger,
	indexManager api.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
) Interface {
	collector := colly.NewCollector(
		colly.MaxDepth(3),
		colly.Async(true),
		colly.Debugger(debugger),
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
		Debugger:         debugger,
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

	// Let Colly handle link discovery
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	return crawler
}
