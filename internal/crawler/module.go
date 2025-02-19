package crawler

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// Module provides the crawler module and its dependencies
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		NewCrawler,
	),
)

// Params holds the dependencies required to create a new Crawler instance
type Params struct {
	fx.In

	Logger   logger.Interface
	Storage  storage.Interface
	Debugger *logger.CollyDebugger
	Config   *config.Config
}

// Result holds the result of creating a new Crawler instance
type Result struct {
	fx.Out

	Crawler *Crawler
}

// Crawler represents a web crawler
type Crawler struct {
	Storage        storage.Interface
	Collector      *colly.Collector
	Logger         logger.Interface
	Debugger       *logger.CollyDebugger
	IndexName      string
	articleChan    chan *models.Article
	ArticleService article.Interface
	IndexSvc       storage.IndexServiceInterface
	Config         *config.Config
}

const (
	DefaultMaxDepth    = 3
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultParallelism = 2
	DefaultBatchSize   = 100
)

// NewCrawler creates a new Crawler instance
func NewCrawler(p Params) (Result, error) {
	if p.Logger == nil {
		return Result{}, errors.New("logger is required")
	}

	if p.Config == nil {
		return Result{}, errors.New("config is required")
	}

	// Log the entire configuration to ensure it's set correctly
	p.Logger.Debug("Initializing Crawler Configuration", "config", p.Config)

	p.Logger.Info("Crawler initialized",
		"maxDepth", p.Config.Crawler.MaxDepth,
		"rateLimit", p.Config.Crawler.RateLimit,
	)

	// Use the collector's New function to create a collector instance
	collectorResult, err := collector.New(collector.Params{
		BaseURL:   p.Config.Crawler.BaseURL,
		MaxDepth:  p.Config.Crawler.MaxDepth,
		RateLimit: p.Config.Crawler.RateLimit,
		Debugger:  p.Debugger,
		Logger:    p.Logger,
	})
	if err != nil {
		return Result{}, err
	}

	crawler := &Crawler{
		Storage:        p.Storage,
		Collector:      collectorResult.Collector,
		Logger:         p.Logger,
		Debugger:       p.Debugger,
		IndexName:      p.Config.Crawler.IndexName,
		articleChan:    make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewService(p.Logger),
		IndexSvc:       storage.NewIndexService(p.Logger),
		Config:         p.Config,
	}

	// Configure collector callbacks
	configureCollectorCallbacks(collectorResult.Collector, crawler)

	return Result{Crawler: crawler}, nil
}

// createCollector initializes a new Colly collector with the specified configuration
func createCollector(config config.CrawlerConfig, log logger.Interface) (*colly.Collector, error) {
	// Log the base URL before parsing
	log.Debug("Creating collector with base URL", "baseURL", config.BaseURL)

	// Parse domain from BaseURL
	parsedURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	domain := parsedURL.Host

	// Check if the domain is empty
	if domain == "" {
		return nil, errors.New("parsed domain is empty; please provide a valid base URL")
	}

	// Allow the main domain and its subdomains
	allowedDomains := []string{domain}
	if strings.HasPrefix(domain, "www.") {
		allowedDomains = append(allowedDomains, strings.TrimPrefix(domain, "www."))
	} else {
		allowedDomains = append(allowedDomains, "www."+domain)
	}

	// Log the allowed domains and other configuration values
	log.Debug("Allowed domains for collector", "allowedDomains", allowedDomains)

	maxDepth := config.MaxDepth
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	// Create a new collector with proper configuration
	c := colly.NewCollector(
		colly.MaxDepth(maxDepth),
		colly.Async(true),
		colly.AllowedDomains(allowedDomains...), // Use the dynamically created allowed domains
		colly.MaxBodySize(DefaultMaxBodySize),
		colly.UserAgent(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
				"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		),
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: config.RateLimit,
		Parallelism: DefaultParallelism,
	})
	if err != nil {
		return nil, fmt.Errorf("error setting rate limit: %w", err)
	}

	return c, nil
}

// configureCollectorCallbacks sets up the callbacks for the Colly collector
func configureCollectorCallbacks(c *colly.Collector, crawler *Crawler) {
	c.OnRequest(func(r *colly.Request) {
		crawler.Logger.Debug("Requesting URL", "url", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.Logger.Debug("Received response", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		crawler.Logger.Error("Error scraping", "url", r.Request.URL.String(), "error", err)
	})

	c.OnHTML("div.details", func(e *colly.HTMLElement) {
		crawler.Logger.Debug("Found details", "url", e.Request.URL.String())
		crawler.ProcessPage(e)
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
