// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common/transport"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultChannelBufferSize is the default buffer size for processor channels.
	DefaultChannelBufferSize = 100
	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 100
	// DefaultMaxIdleConnsPerHost is the default maximum number of idle connections per host.
	DefaultMaxIdleConnsPerHost = 10
	// DefaultIdleConnTimeout is the default idle connection timeout.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultResponseHeaderTimeout is the default response header timeout.
	DefaultResponseHeaderTimeout = 30 * time.Second
	// DefaultExpectContinueTimeout is the default expect continue timeout.
	DefaultExpectContinueTimeout = 1 * time.Second
)

// CrawlerParams holds parameters for creating a crawler instance
type CrawlerParams struct {
	Logger         logger.Interface
	Bus            *events.EventBus
	IndexManager   types.IndexManager
	Sources        sources.Interface
	Config         *crawler.Config
	ArticleService articles.Interface
	PageService    page.Interface
	Storage        types.Interface
}

// CrawlerResult holds the crawler instance and its channels
type CrawlerResult struct {
	Crawler        Interface
	ArticleChannel chan *domain.Article
	PageChannel    chan *domain.Page
}

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
func createCollector(cfg *crawler.Config, log logger.Interface) (*colly.Collector, error) {
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
		log.Warn("TLS certificate verification is disabled. This is not recommended for production use.",
			"component", "crawler",
			"warning", "This makes HTTPS connections vulnerable to man-in-the-middle attacks")
	}

	// Set up callbacks
	collector.OnRequest(func(r *colly.Request) {
		log.Info("Visiting", "url", r.URL.String())
	})

	collector.OnResponse(func(r *colly.Response) {
		log.Info("Visited", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", err)
	})

	return collector, nil
}

// NewCrawlerWithParams creates a new crawler instance with all its components.
// This is the non-FX version that replaces ProvideCrawler.
func NewCrawlerWithParams(p CrawlerParams) (*CrawlerResult, error) {
	validator := createJobValidator()

	// Create channels
	articleChannel := make(chan *domain.Article, ArticleChannelBufferSize)
	pageChannel := make(chan *domain.Page, DefaultChannelBufferSize)

	// Create processors
	articleProcessor := articles.NewProcessor(
		p.Logger,
		p.ArticleService,
		validator,
		p.Storage,
		"articles",
		articleChannel,
		nil,
		nil,
	)

	pageProcessor := page.NewPageProcessor(
		p.Logger,
		p.PageService,
		validator,
		p.Storage,
		"pages",
		pageChannel,
	)

	// Create collector
	collector, err := createCollector(p.Config, p.Logger)
	if err != nil {
		return nil, err
	}

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
		htmlProcessor:    NewHTMLProcessor(p.Logger, p.Sources),
		cfg:              p.Config,
		abortChan:        make(chan struct{}),
	}

	c.linkHandler = NewLinkHandler(c)

	return &CrawlerResult{
		Crawler:        c,
		ArticleChannel: articleChannel,
		PageChannel:    pageChannel,
	}, nil
}
