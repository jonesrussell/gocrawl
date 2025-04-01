// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/pkg/collector"
)

// Crawler implements the crawler Interface.
type Crawler struct {
	collector        *colly.Collector
	Logger           common.Logger
	Debugger         debug.Debugger
	bus              *events.Bus
	indexManager     api.IndexManager
	sources          sources.Interface
	articleProcessor collector.Processor
	contentProcessor collector.Processor
}

var _ Interface = (*Crawler)(nil)

// Start begins crawling from the given base URL.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	c.Logger.Info("Starting crawler", "source", sourceName)

	// Get source configuration
	source, err := c.sources.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("failed to get source configuration: %w", err)
	}

	// Create collector params
	params := collector.Params{
		BaseURL:     source.URL,
		MaxDepth:    source.MaxDepth,
		RateLimit:   source.RateLimit,
		Logger:      c.Logger,
		Context:     ctx,
		Done:        make(chan struct{}),
		Parallelism: 2, // Set a default parallelism value
		Source: &config.Source{
			URL:       source.URL,
			MaxDepth:  source.MaxDepth,
			RateLimit: source.RateLimit,
		},
		ArticleProcessor: c.articleProcessor,
		ContentProcessor: c.contentProcessor,
	}

	// Create collector
	result, err := collector.New(params)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}

	// Set the new collector
	c.collector = result.Collector

	// Start crawling
	return c.collector.Visit(source.URL)
}

// Stop gracefully stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	c.Logger.Info("Stopping crawler")
	c.collector.Wait()
	return nil
}

// Subscribe adds a content handler to receive discovered content.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: duration,
		Parallelism: 1,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}
	return nil
}

// SetMaxDepth sets the maximum crawl depth.
func (c *Crawler) SetMaxDepth(depth int) {
	c.collector.MaxDepth = depth
}

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.collector = collector
}

// GetIndexManager returns the index manager interface.
func (c *Crawler) GetIndexManager() api.IndexManager {
	return c.indexManager
}

// Wait blocks until the crawler has finished processing all queued requests.
func (c *Crawler) Wait() {
	c.collector.Wait()
}

// GetMetrics returns the current crawler metrics.
func (c *Crawler) GetMetrics() *collector.Metrics {
	return &collector.Metrics{
		PagesVisited:  int64(c.collector.ID),
		ArticlesFound: 0,
		Errors:        0,
		StartTime:     time.Now().Unix(),
	}
}
