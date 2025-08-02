// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"fmt"
	"net/http"
	"time"

	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common/transport"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
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

// ModuleParams contains dependencies for creating the crawler module.
type ModuleParams struct {
	fx.In

	Config  config.Interface
	Logger  logger.Interface
	Storage types.Interface
}

// Module provides the crawler module's dependencies.
var Module = fx.Module("crawler",
	fx.Provide(
		ProvideCrawler,
		ProvideProcessorFactory,
	),
)

// createJobValidator creates a simple job validator
func createJobValidator() content.JobValidator {
	return &struct {
		content.JobValidator
	}{
		JobValidator: content.JobValidatorFunc(func(job *content.Job) error {
			if job == nil {
				return errors.New("job cannot be nil")
			}
			if job.URL == "" {
				return errors.New("job URL cannot be empty")
			}
			return nil
		}),
	}
}

// createCollector creates and configures a new colly collector
func createCollector(cfg *crawler.Config, logger logger.Interface) (*colly.Collector, error) {
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
		return nil, fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Configure transport
	tlsConfig, err := transport.NewTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
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

	return collector, nil
}

// ProvideCrawler creates a new crawler instance with all its components
func ProvideCrawler(p struct {
	fx.In

	Logger         logger.Interface
	Bus            *events.EventBus
	IndexManager   types.IndexManager
	Sources        sources.Interface
	Config         *crawler.Config
	ArticleService articles.Interface
	PageService    page.Interface
	Storage        types.Interface
}) (struct {
	fx.Out

	Crawler        Interface
	ArticleChannel chan *models.Article
	PageChannel    chan *models.Page
}, error) {
	validator := createJobValidator()

	// Create processors
	articleProcessor := articles.NewProcessor(
		p.Logger,
		p.ArticleService,
		validator,
		p.Storage,
		"articles",
		make(chan *models.Article, ArticleChannelBufferSize),
		nil,
		nil,
	)

	pageProcessor := page.NewPageProcessor(
		p.Logger,
		p.PageService,
		validator,
		p.Storage,
		"pages",
		make(chan *models.Page, DefaultChannelBufferSize),
	)

	// Create collector
	collector, err := createCollector(p.Config, p.Logger)
	if err != nil {
		return struct {
			fx.Out
			Crawler        Interface
			ArticleChannel chan *models.Article
			PageChannel    chan *models.Page
		}{}, err
	}

	// Create channels
	articleChannel := make(chan *models.Article, ArticleChannelBufferSize)
	pageChannel := make(chan *models.Page, DefaultChannelBufferSize)

	// Create crawler
	c := &Crawler{
		logger:           p.Logger,
		collector:        collector,
		bus:              p.Bus,
		indexManager:     p.IndexManager,
		sources:          p.Sources,
		articleProcessor: articleProcessor,
		pageProcessor:    pageProcessor,
		state:            NewState(p.Logger),
		done:             make(chan struct{}),
		articleChannel:   articleChannel,
		processors:       []content.Processor{articleProcessor, pageProcessor},
		htmlProcessor:    NewHTMLProcessor(p.Logger),
		cfg:              p.Config,
		abortChan:        make(chan struct{}),
	}

	c.linkHandler = NewLinkHandler(c)

	return struct {
		fx.Out
		Crawler        Interface
		ArticleChannel chan *models.Article
		PageChannel    chan *models.Page
	}{
		Crawler:        c,
		ArticleChannel: articleChannel,
		PageChannel:    pageChannel,
	}, nil
}

// ProvideProcessorFactory creates a new processor factory.
func ProvideProcessorFactory(p ModuleParams) ProcessorFactory {
	return NewProcessorFactory(p.Logger, p.Storage, "content")
}
