// Package crawler provides the core crawling functionality for GoCrawl.
// It manages the crawling process, including URL processing, rate limiting,
// and content extraction.
package crawler

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
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

// ProvideCrawler creates a new crawler instance with the given dependencies.
func ProvideCrawler(
	logger logger.Interface,
	indexManager interfaces.IndexManager,
	sources sources.Interface,
	processors []common.Processor,
	bus *events.EventBus,
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
		bus,
		indexManager,
		sources,
		articleProcessor,
		contentProcessor,
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
		// Provide the crawler
		ProvideCrawler,
	),
)

// NewCrawler creates a new crawler instance.
func NewCrawler(
	logger logger.Interface,
	bus *events.EventBus,
	indexManager interfaces.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	cfg *crawler.Config,
) Interface {
	collector := colly.NewCollector(
		colly.MaxDepth(cfg.MaxDepth),
		colly.Async(true),
		colly.AllowedDomains(cfg.AllowedDomains...),
		colly.ParseHTTPErrorResponse(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(cfg.UserAgent),
		colly.AllowURLRevisit(),
	)

	// Configure rate limiting
	if err := collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       cfg.Delay,
		RandomDelay: cfg.RandomDelay,
		Parallelism: cfg.MaxConcurrency,
	}); err != nil {
		logger.Error("Failed to set rate limit",
			"error", err,
			"delay", cfg.Delay,
			"randomDelay", cfg.RandomDelay,
			"parallelism", cfg.MaxConcurrency)
	}

	// Configure transport with more reasonable settings
	collector.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	// Set up callbacks
	collector.OnRequest(func(r *colly.Request) {
		logger.Info("Visiting", "url", r.URL.String())
	})

	collector.OnResponse(func(r *colly.Response) {
		logger.Info("Visited", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	collector.OnError(func(r *colly.Response, err error) {
		logger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", err)
	})

	c := &Crawler{
		logger:           logger,
		collector:        collector,
		bus:              bus,
		indexManager:     indexManager,
		sources:          sources,
		articleProcessor: articleProcessor,
		contentProcessor: contentProcessor,
		state:            NewState(logger),
		done:             make(chan struct{}),
		articleChannel:   make(chan *models.Article, ArticleChannelBufferSize),
		processors:       make([]common.Processor, 0),
		cfg:              cfg,
	}

	c.linkHandler = NewLinkHandler(c)
	c.htmlProcessor = NewHTMLProcessor(c)

	return c
}
