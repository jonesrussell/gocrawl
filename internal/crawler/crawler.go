// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

// Crawler represents a web crawler instance.
type Crawler struct {
	collector    *colly.Collector
	Logger       common.Logger
	Debugger     debug.Debugger
	bus          *events.Bus
	baseURL      string
	ctx          context.Context
	cancel       context.CancelFunc
	indexManager api.IndexManager
	sources      sources.Interface
	done         chan struct{}
	mu           sync.RWMutex
	metrics      *collector.Metrics
}

var _ Interface = (*Crawler)(nil)

// Start begins the crawling process at the specified base URL.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the source configuration
	source, err := c.sources.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("failed to find source %s: %w", sourceName, err)
	}

	if source.URL == "" {
		return errors.New("source URL cannot be empty")
	}

	// Initialize crawler state
	c.baseURL = source.URL
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.done = make(chan struct{})
	c.metrics = collector.NewMetrics()

	// Ensure collector is initialized
	if c.collector == nil {
		return errors.New("collector not initialized")
	}

	// Configure collector based on source
	params := collector.Params{
		BaseURL:   source.URL,
		MaxDepth:  source.MaxDepth,
		RateLimit: source.RateLimit,
		Logger:    c.Logger,
		Context:   c.ctx,
		Done:      c.done,
		Source: &config.Source{
			URL:       source.URL,
			MaxDepth:  source.MaxDepth,
			RateLimit: source.RateLimit,
		},
	}
	if c.Debugger != nil {
		params.Debugger = &logger.CollyDebugger{
			Logger: c.Logger,
		}
	}

	result, err := collector.New(params)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}
	c.collector = result.Collector

	c.Logger.Info("Starting crawl",
		"url", c.baseURL,
		"max_depth", source.MaxDepth,
		"rate_limit", source.RateLimit)

	c.setupCallbacks()

	// Start crawling
	if visitErr := c.collector.Visit(c.baseURL); visitErr != nil {
		c.Logger.Error("Error visiting base URL", "url", c.baseURL, "error", visitErr)
		return fmt.Errorf("error starting crawl for base URL %s: %w", c.baseURL, visitErr)
	}

	// Monitor context for cancellation while waiting
	go c.monitorContext(c.ctx)

	// Wait for crawling to complete
	go func() {
		c.collector.Wait()
		close(c.done)
		c.Logger.Info("Crawling completed",
			"pages_visited", c.metrics.PagesVisited,
			"articles_found", c.metrics.ArticlesFound,
			"errors", c.metrics.Errors,
			"duration", time.Since(time.Unix(c.metrics.StartTime, 0)))
	}()

	return nil
}

// Stop gracefully stops the crawler, respecting the provided context.
func (c *Crawler) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	if c.collector != nil {
		c.collector.Wait()
	}

	return nil
}

// Subscribe adds a content handler to receive discovered content.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if c.collector == nil {
		return errors.New("collector not initialized")
	}
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
	if c.collector != nil {
		c.collector.MaxDepth = depth
	}
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
	if c.collector != nil {
		c.collector.Wait()
	}
}

// GetMetrics returns the current crawler metrics.
func (c *Crawler) GetMetrics() *collector.Metrics {
	return c.metrics
}

// monitorContext monitors the context for cancellation.
func (c *Crawler) monitorContext(ctx context.Context) {
	<-ctx.Done()
	if err := c.Stop(ctx); err != nil {
		c.Logger.Error("Failed to stop crawler", "error", err)
	}
}

// setupCallbacks sets up the crawler's event callbacks.
func (c *Crawler) setupCallbacks() {
	if c.collector == nil {
		return
	}

	c.collector.OnRequest(func(r *colly.Request) {
		// No need to update LastRequestTime as it's not part of the metrics struct
	})

	c.collector.OnResponse(func(r *colly.Response) {
		c.metrics.PagesVisited++
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		c.metrics.Errors++
		c.Logger.Error("Request error",
			"url", r.Request.URL.String(),
			"status_code", r.StatusCode,
			"error", err)
	})
}
