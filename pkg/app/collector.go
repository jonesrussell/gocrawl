// Package app provides application-level utilities and setup functions.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/jonesrussell/gocrawl/pkg/config"
)

// CollectorResult holds the result of setting up a collector.
type CollectorResult struct {
	Collector *colly.Collector
	Crawler   crawler.Interface
}

// SetupCollector creates and configures a new collector instance.
func SetupCollector(
	ctx context.Context,
	log logger.Interface,
	source sources.Config,
	processors []collector.Processor,
	done chan struct{},
	cfg config.Interface,
) (*CollectorResult, error) {
	if source.URL == "" {
		return nil, errors.New("source URL is required")
	}

	if len(processors) == 0 {
		return nil, errors.New("at least one processor is required")
	}

	// Create new collector with rate limiting
	c := colly.NewCollector(
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
	)

	// Set rate limiting
	if source.RateLimit > 0 {
		if err := c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			RandomDelay: source.RateLimit,
			Parallelism: config.DefaultParallelism,
		}); err != nil {
			return nil, fmt.Errorf("failed to set rate limit: %w", err)
		}
	}

	// Create event bus
	bus := events.NewBus()

	// Create sources instance
	sources := sources.NewSources(&source, log)

	// Create mock index manager for testing
	indexManager := &api.MockIndexManager{}

	// Find article and content processors
	var articleProc, contentProc collector.Processor
	for _, p := range processors {
		switch p.(type) {
		case *article.ArticleProcessor:
			articleProc = p
		case *content.ContentProcessor:
			contentProc = p
		}
	}

	if articleProc == nil {
		return nil, errors.New("article processor is required")
	}
	if contentProc == nil {
		return nil, errors.New("content processor is required")
	}

	// Create crawler with dependencies
	crawlerResult := crawler.ProvideCrawler(
		log,
		logger.NewCollyDebugger(log),
		indexManager,
		sources,
		articleProc,
		contentProc,
		bus,
	)

	// Add processors
	for _, p := range processors {
		c.OnHTML("*", func(e *colly.HTMLElement) {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			default:
				if processErr := p.Process(e); processErr != nil {
					crawlerResult.Crawler.GetMetrics().Errors++
					log.Error("Failed to process HTML element", "error", processErr)
				}
			}
		})
	}

	// Add error handler
	c.OnError(func(r *colly.Response, err error) {
		crawlerResult.Crawler.GetMetrics().Errors++
		log.Error("Collector error", "url", r.Request.URL, "error", err)
	})

	return &CollectorResult{
		Collector: c,
		Crawler:   crawlerResult.Crawler,
	}, nil
}

// ConfigureCrawler configures a crawler instance with the given source.
func ConfigureCrawler(
	c crawler.Interface,
	source sources.Config,
	collectorResult *CollectorResult,
) error {
	if source.URL == "" {
		return errors.New("source URL is required")
	}

	// Set collector
	c.SetCollector(collectorResult.Collector)

	return nil
}

// NewCollector creates a new collector with the given configuration.
func NewCollector(cfg common.Config) (*colly.Collector, error) {
	ctx := context.Background()
	log := logger.NewNoOp()

	sources := cfg.GetSources()
	if len(sources) == 0 {
		return nil, errors.New("no sources configured")
	}
	source := &sources[0] // Use the first source

	params := collector.Params{
		BaseURL:     source.URL,
		MaxDepth:    source.MaxDepth,
		RateLimit:   source.RateLimit,
		Logger:      log,
		Context:     ctx,
		Done:        make(chan struct{}),
		Parallelism: config.DefaultParallelism,
		Source:      source,
	}

	result, err := collector.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create collector: %w", err)
	}

	return result.Collector, nil
}
