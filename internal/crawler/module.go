// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/transport"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
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
	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 100
	// DefaultMaxIdleConnsPerHost is the default maximum number of idle connections per host.
	DefaultMaxIdleConnsPerHost = 10
	// DefaultIdleConnTimeout is the default idle connection timeout.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultTLSHandshakeTimeout is the default TLS handshake timeout.
	DefaultTLSHandshakeTimeout = 10 * time.Second
	// DefaultResponseHeaderTimeout is the default response header timeout.
	DefaultResponseHeaderTimeout = 30 * time.Second
	// DefaultExpectContinueTimeout is the default expect continue timeout.
	DefaultExpectContinueTimeout = 1 * time.Second
	// DefaultMaxRetries is the default maximum number of retries for failed requests.
	DefaultMaxRetries = 3
	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 1 * time.Second
	// DefaultMaxDepth is the default maximum depth for crawling.
	DefaultMaxDepth = 1
	// DefaultMaxBodySize is the default maximum body size for responses.
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	// DefaultMaxConcurrentRequests is the default maximum number of concurrent requests.
	DefaultMaxConcurrentRequests = 10
	// DefaultRequestTimeout is the default timeout for requests.
	DefaultRequestTimeout = 30 * time.Second
	// DefaultUserAgent is the default user agent string used for HTTP requests.
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/91.0.4472.124 Safari/537.36"
	// DefaultChannelBufferSize is the default buffer size for processor channels.
	DefaultChannelBufferSize = 100
)

// CrawlerParams defines the parameters for creating a crawler.
type CrawlerParams struct {
	fx.In
	Logger       logger.Interface
	Bus          *events.EventBus
	IndexManager interfaces.IndexManager
	Sources      sources.Interface
	Config       *crawler.Config
}

// Result defines the crawler module's output.
type Result struct {
	fx.Out
	Crawler Interface
}

// ProcessorFactory creates processors for the crawler.
type ProcessorFactory interface {
	CreateProcessors(validator common.Processor) ([]common.Processor, error)
}

// DefaultProcessorFactory implements ProcessorFactory.
type DefaultProcessorFactory struct {
	logger logger.Interface
}

// NewProcessorFactory creates a new processor factory.
func NewProcessorFactory(logger logger.Interface) *DefaultProcessorFactory {
	return &DefaultProcessorFactory{
		logger: logger,
	}
}

// CreateProcessors creates processors for the crawler.
func (f *DefaultProcessorFactory) CreateProcessors(validator common.Processor) ([]common.Processor, error) {
	processors := make([]common.Processor, 0, 2) // Pre-allocate for 2 processors

	// Create article processor
	articleProcessor := articles.NewProcessor(articles.ProcessorParams{
		Logger:    f.logger,
		Validator: validator,
	})
	processors = append(processors, articleProcessor)

	// Create page processor
	pageProcessor := page.NewPageProcessor(page.ProcessorParams{
		Logger:    f.logger,
		Validator: validator,
	})
	processors = append(processors, pageProcessor)

	return processors, nil
}

// ProvideCrawler provides a crawler instance.
func ProvideCrawler(
	params CrawlerParams,
	processors []common.Processor,
) (Interface, error) {
	var articleProcessor, pageProcessor common.Processor

	// Find article and page processors
	for _, p := range processors {
		if p.ContentType() == content.ContentTypeArticle {
			articleProcessor = p
		} else if p.ContentType() == content.ContentTypePage {
			pageProcessor = p
		}
	}

	if articleProcessor == nil || pageProcessor == nil {
		return nil, fmt.Errorf("missing required processors")
	}

	return NewCrawler(
		params.Logger,
		params.Bus,
		params.IndexManager,
		params.Sources,
		articleProcessor,
		pageProcessor,
		params.Config,
	), nil
}

// Module provides the crawler module for dependency injection.
var Module = fx.Module("crawler",
	fx.Provide(
		NewProcessorFactory,
		ProvideCrawler,
	),
	fx.Invoke(func(c Interface) {
		// Initialize the crawler
	}),
)

// NewCrawler creates a new crawler instance.
func NewCrawler(
	logger logger.Interface,
	bus *events.EventBus,
	indexManager interfaces.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	pageProcessor common.Processor,
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
	tlsConfig, err := transport.NewTLSConfig(cfg)
	if err != nil {
		logger.Error("Failed to create TLS configuration",
			"error", err)
		return nil
	}

	collector.WithTransport(&http.Transport{
		TLSClientConfig:       tlsConfig,
		DisableKeepAlives:     false,
		MaxIdleConns:          DefaultMaxIdleConns,
		MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		ResponseHeaderTimeout: DefaultResponseHeaderTimeout,
		ExpectContinueTimeout: DefaultExpectContinueTimeout,
	})

	if cfg.TLS.InsecureSkipVerify {
		logger.Warn("TLS certificate verification is disabled. This is not recommended for production use.",
			"component", "crawler",
			"warning", "This makes HTTPS connections vulnerable to man-in-the-middle attacks")
	}

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
		pageProcessor:    pageProcessor,
		state:            NewState(logger),
		done:             make(chan struct{}),
		articleChannel:   make(chan *models.Article, ArticleChannelBufferSize),
		processors:       make([]common.Processor, 0),
		htmlProcessor:    NewHTMLProcessor(logger),
		cfg:              cfg,
		abortChan:        make(chan struct{}),
	}

	c.linkHandler = NewLinkHandler(c)

	return c
}
