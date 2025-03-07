// Package crawler provides the core crawling functionality for the application.
// This file contains the dependency injection module configuration.
package crawler

import (
	"context"
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Interface defines the methods required for a crawler.
// It provides the core functionality for web crawling operations.
type Interface interface {
	// Start begins the crawling process at the specified URL.
	// It manages the crawling lifecycle and handles errors.
	Start(ctx context.Context, url string) error
	// Stop performs cleanup operations when the crawler is stopped.
	Stop()
	// SetCollector sets the collector for the crawler.
	// This allows for dependency injection and testing.
	SetCollector(collector *colly.Collector)
	// SetService sets the article service for the crawler.
	// This allows for dependency injection and testing.
	SetService(service article.Interface)
	// GetBaseURL returns the base URL from the configuration.
	GetBaseURL() string
	// GetIndexManager returns the index service interface.
	GetIndexManager() storage.IndexServiceInterface
}

// provideCollyDebugger creates a new CollyDebugger instance for debugging collector operations.
// It takes a logger interface and returns a configured debugger.
func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// Params holds the dependencies for creating a crawler.
// It uses fx.In for dependency injection.
type Params struct {
	fx.In

	// Logger provides structured logging capabilities
	Logger logger.Interface
	// Storage handles content storage operations
	Storage storage.Interface
	// Debugger handles debugging operations
	Debugger *logger.CollyDebugger
	// Config holds the crawler configuration
	Config *config.Config
	// Source is the name of the source being crawled
	Source string `name:"sourceName"`
	// IndexService manages index operations
	IndexService storage.IndexServiceInterface
	// ContentProcessor handles content processing operations
	ContentProcessor []models.ContentProcessor `group:"processors"`
}

// Result holds the crawler instance.
// It uses fx.Out for dependency injection.
type Result struct {
	fx.Out

	// Crawler is the crawler interface implementation
	Crawler Interface
}

// Module provides the dependency injection configuration for the crawler package.
// It exports the crawler interface and provides implementations for all required
// components including the crawler instance, collector, and related services.
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		ProvideCrawler,
	),
)

// ProvideCrawler creates a new Crawler instance with all required dependencies.
// It validates the dependencies and initializes the crawler with the provided configuration.
func ProvideCrawler(p Params) (Interface, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	if p.Config == nil {
		return nil, errors.New("config is required")
	}

	if p.Storage == nil {
		return nil, errors.New("storage is required")
	}

	if p.IndexService == nil {
		return nil, errors.New("index service is required")
	}

	if len(p.ContentProcessor) == 0 {
		return nil, errors.New("at least one content processor is required")
	}

	// Log the entire configuration to ensure it's set correctly
	p.Logger.Debug("Initializing Crawler Configuration", "config", p.Config)

	// Create a new crawler instance
	crawler := &Crawler{
		Storage:     p.Storage,
		Logger:      p.Logger,
		Debugger:    p.Debugger,
		IndexName:   p.Config.Crawler.IndexName,
		articleChan: make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewServiceWithConfig(article.ServiceParams{
			Logger: p.Logger,
			Config: p.Config,
			Source: p.Source,
		}),
		IndexService:     p.IndexService,
		Config:           p.Config,
		ContentProcessor: p.ContentProcessor[0], // Use the first processor
	}

	return crawler, nil
}

const (
	// DefaultBatchSize is the default size for buffered channels used for
	// processing articles during crawling.
	DefaultBatchSize = 100
)

// NewCollector creates a new collector instance with the specified configuration.
// It sets up rate limiting, parallelism, and other collector-specific settings.
func NewCollector(cfg *config.Config) *colly.Collector {
	// Create a new collector with the specified configuration
	c := colly.NewCollector(
		colly.MaxDepth(cfg.Crawler.MaxDepth),
		colly.Async(true),
	)

	// Set up rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: cfg.Crawler.RandomDelay,
		Parallelism: cfg.Crawler.Parallelism,
	})

	return c
}

// New creates a new crawler instance with all required dependencies.
// It initializes the crawler with the provided configuration and services.
func New(
	storage storage.Interface,
	collector *colly.Collector,
	logger logger.Interface,
	articleService article.Interface,
	indexService storage.IndexServiceInterface,
	cfg *config.Config,
) Interface {
	// Create a new crawler instance
	c := &Crawler{
		Storage:        storage,
		Collector:      collector,
		Logger:         logger,
		IndexName:      cfg.Elasticsearch.IndexName,
		articleChan:    make(chan *models.Article, 100),
		ArticleService: articleService,
		IndexService:   indexService,
		Config:         cfg,
	}

	// Set up the debugger if debug mode is enabled
	if cfg.Log.Debug {
		c.Debugger = provideCollyDebugger(logger)
		c.Collector.SetDebugger(c.Debugger)
	}

	return c
}
