// Package app provides application-level utilities and setup functions.
package app

import (
	"context"
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/pkg/config"
	"github.com/jonesrussell/gocrawl/pkg/logger"
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
		return nil, fmt.Errorf("source URL is required")
	}

	if len(processors) == 0 {
		return nil, fmt.Errorf("at least one processor is required")
	}

	// Create new collector with rate limiting
	c := colly.NewCollector(
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
	)

	// Set rate limiting
	if source.RateLimit > 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			RandomDelay: source.RateLimit,
			Parallelism: 1,
		})
	}

	// Create event bus
	bus := events.NewBus()

	// Create sources instance
	sources := sources.NewSources(&source)

	// Create mock index manager for testing
	indexManager := &api.MockIndexManager{}

	// Create crawler with dependencies
	crawlerResult, err := crawler.ProvideCrawler(crawler.Params{
		Logger:       log,
		Sources:      sources,
		IndexManager: indexManager,
	}, bus)
	if err != nil {
		return nil, fmt.Errorf("failed to create crawler: %w", err)
	}

	// Add processors
	for _, p := range processors {
		c.OnHTML("*", func(e *colly.HTMLElement) {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			default:
				if err := p.Process(e); err != nil {
					crawlerResult.Crawler.GetMetrics().Errors++
					log.Error("Failed to process HTML element", "error", err)
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
		return fmt.Errorf("source URL is required")
	}

	// Set collector
	c.SetCollector(collectorResult.Collector)

	return nil
}
