package crawler

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
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

	p.Logger.Info("Crawler initialized",
		"baseURL", p.Config.Crawler.BaseURL,
		"maxDepth", p.Config.Crawler.MaxDepth,
		"rateLimit", p.Config.Crawler.RateLimit,
	)

	collector, err := createCollector(p.Config.Crawler)
	if err != nil {
		return Result{}, err
	}

	crawler := &Crawler{
		Storage:        p.Storage,
		Collector:      collector,
		Logger:         p.Logger,
		Debugger:       p.Debugger,
		IndexName:      p.Config.Crawler.IndexName,
		articleChan:    make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewService(p.Logger),
		IndexSvc:       storage.NewIndexService(p.Logger),
		Config:         p.Config,
	}

	// Configure collector callbacks
	configureCollectorCallbacks(collector, crawler)

	return Result{Crawler: crawler}, nil
}

func createCollector(config config.CrawlerConfig) (*colly.Collector, error) {
	// Parse domain from BaseURL
	parsedURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	domain := parsedURL.Host

	maxDepth := config.MaxDepth
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
		RandomDelay: config.RateLimit,
		Parallelism: DefaultParallelism,
	})
	if err != nil {
		return nil, fmt.Errorf("error setting rate limit: %w", err)
	}

	return c, nil
}
