package crawler

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
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
