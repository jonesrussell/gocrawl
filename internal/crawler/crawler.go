package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Constants for configuration
const (
	TimeoutDuration  = 5 * time.Second
	HTTPStatusOK     = 200
	DefaultBatchSize = 100
	DefaultMaxDepth  = 2
	DefaultRateLimit = time.Second
)

// Constants for default values
const (
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultParallelism = 2                // Default parallelism
)

// Crawler struct to hold configuration or state if needed
type Crawler struct {
	Storage        storage.Interface
	Collector      *colly.Collector
	Logger         logger.Interface
	Debugger       *logger.CollyDebugger
	IndexName      string
	articleChan    chan *models.Article
	ArticleService article.Interface
	indexSvc       *storage.IndexService
	Config         *config.Config // Rename to Config for clarity
}

// Params holds the parameters for creating a Crawler
type Params struct {
	fx.In

	Config   *config.Config
	Debugger *logger.CollyDebugger
	Logger   logger.Interface
	Storage  storage.Interface
}

// Result holds the dependencies for the crawler
type Result struct {
	fx.Out

	Crawler *Crawler
}

// NewCrawler creates a new Crawler instance
func NewCrawler(p Params) (Result, error) {
	if p.Logger == nil {
		return Result{}, errors.New("logger is required")
	}

	// Log the logger being used
	p.Logger.Info("Creating new Crawler instance")

	// Parse domain from BaseURL
	parsedURL, err := url.Parse(p.Config.Crawler.BaseURL)
	if err != nil {
		return Result{}, fmt.Errorf("invalid base URL: %w", err)
	}
	domain := parsedURL.Host

	maxDepth := p.Config.Crawler.MaxDepth
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	// Create a new collector with proper configuration
	c := colly.NewCollector(
		colly.MaxDepth(maxDepth),
		colly.Async(true),
		colly.AllowedDomains(domain),
		colly.MaxBodySize(DefaultMaxBodySize),
		colly.UserAgent(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
				"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		),
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.Config.Crawler.RateLimit,
		Parallelism: DefaultParallelism,
	})
	if err != nil {
		return Result{}, fmt.Errorf("error setting rate limit: %w", err)
	}

	crawler := &Crawler{
		Storage:        p.Storage,
		Collector:      c,
		Logger:         p.Logger,
		Debugger:       p.Debugger,
		IndexName:      p.Config.Crawler.IndexName,
		articleChan:    make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewService(p.Logger),
		indexSvc:       storage.NewIndexService(p.Storage, p.Logger),
		Config:         p.Config,
	}

	// Configure collector callbacks
	configureCollectorCallbacks(c, crawler)

	p.Logger.Info("Crawler initialized",
		"baseURL", p.Config.Crawler.BaseURL,
		"maxDepth", maxDepth,
		"rateLimit", p.Config.Crawler.RateLimit,
		"domain", domain)

	return Result{Crawler: crawler}, nil
}

// Start method to begin crawling
func (c *Crawler) Start(ctx context.Context) error {
	// Log full crawler configuration when debug is enabled
	if c.Config.App.Debug {
		c.Logger.Debug("Starting crawler with configuration",
			"baseURL", c.Config.Crawler.BaseURL,
			"maxDepth", c.Config.Crawler.MaxDepth,
			"rateLimit", c.Config.Crawler.RateLimit,
			"indexName", c.Config.Crawler.IndexName,
			"skipTLS", c.Config.Crawler.SkipTLS,
			"environment", c.Config.App.Environment,
			"logLevel", c.Config.App.LogLevel,
			"elasticURL", c.Config.Elasticsearch.URL,
		)
	}

	c.Logger.Debug("Starting crawl at base URL", "url", c.Config.Crawler.BaseURL)

	// Perform initial setup (e.g., test connection, ensure index)
	if err := c.Storage.TestConnection(ctx); err != nil {
		c.Logger.Error("Storage connection failed", "error", err)
		return fmt.Errorf("storage connection failed: %w", err)
	}

	if err := c.indexSvc.EnsureIndex(ctx, c.IndexName); err != nil {
		c.Logger.Error("Index setup failed", "error", err)
		return fmt.Errorf("index setup failed: %w", err)
	}

	// Create a channel to track completion
	done := make(chan struct{})

	// Start crawling in a goroutine
	go func() {
		defer close(done)
		// Visit the base URL to start crawling
		if err := c.Collector.Visit(c.Config.Crawler.BaseURL); err != nil {
			c.Logger.Error("Failed to visit base URL", "error", err)
			return
		}
		// Wait for collector to finish all requests
		c.Collector.Wait()
		c.Logger.Info("Crawler finished - no more links to visit")
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		c.Logger.Info("Crawler stopping due to context cancellation")
		return ctx.Err()
	case <-done:
		c.Logger.Info("Crawler completed successfully")
	}

	return nil
}

// Stop method to cleanly shut down the crawler
func (c *Crawler) Stop() {
	// Perform any necessary cleanup here
}

// processPage handles article extraction
func (c *Crawler) processPage(e *colly.HTMLElement) {
	c.Logger.Debug("Processing page", "url", e.Request.URL.String())
	article := c.ArticleService.ExtractArticle(e)
	if article == nil {
		c.Logger.Debug("No article extracted", "url", e.Request.URL.String())
		return
	}
	c.Logger.Debug("Article extracted", "url", e.Request.URL.String(), "title", article.Title)

	// Index the article after extraction
	if err := c.Storage.IndexDocument(context.Background(), "articles", article.ID, article); err != nil {
		c.Logger.Error("Failed to index article", "articleID", article.ID, "error", err)
		return
	}

	c.articleChan <- article
}

// Add these methods to the Crawler struct
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.Collector = collector
}

func (c *Crawler) SetService(svc article.Interface) {
	c.ArticleService = svc
}

func configureCollectorCallbacks(c *colly.Collector, crawler *Crawler) {
	c.OnRequest(func(r *colly.Request) {
		crawler.Logger.Debug("Requesting URL", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.Logger.Debug("Received response", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error scraping", "url", r.Request.URL.String(), "error", err)
	})

	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		crawler.Logger.Debug("Found details", "url", e.Request.URL.String())
		crawler.processPage(e)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link == "" {
			return
		}
		crawler.Logger.Debug("Found link", "url", link)
		if err := e.Request.Visit(link); err != nil {
			crawler.Logger.Debug("Could not visit link", "url", link, "error", err)
		}
	})

	if crawler.Debugger != nil {
		c.SetDebugger(crawler.Debugger)
	}
}

// Getter methods for configuration
func (c *Crawler) GetBaseURL() string {
	return c.Config.Crawler.BaseURL
}

func (c *Crawler) GetMaxDepth() int {
	return c.Config.Crawler.MaxDepth
}

func (c *Crawler) GetRateLimit() time.Duration {
	return c.Config.Crawler.RateLimit
}
