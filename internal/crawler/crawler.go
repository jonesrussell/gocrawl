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
	metrics      *Metrics
}

// Metrics holds crawler metrics.
type Metrics struct {
	PagesVisited    int64
	ArticlesFound   int64
	Errors          int64
	StartTime       time.Time
	LastRequestTime time.Time
}

// NewMetrics creates a new metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
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
	c.metrics = NewMetrics()

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
			"duration", time.Since(c.metrics.StartTime))
	}()

	return nil
}

// Stop gracefully stops the crawler, respecting the provided context.
func (c *Crawler) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger.Info("Stopping crawler")

	stopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if c.cancel != nil {
		c.cancel()
	}

	select {
	case <-stopCtx.Done():
		c.Logger.Warn("Context for Stop was cancelled", "error", stopCtx.Err())
		return stopCtx.Err()
	case <-c.done:
		c.Logger.Info("Crawler stopped successfully")
		return nil
	}
}

// Wait blocks until the crawler has finished processing all queued requests.
func (c *Crawler) Wait() {
	c.mu.RLock()
	done := c.done
	c.mu.RUnlock()
	if done != nil {
		<-done
	}
}

// Subscribe adds a content handler to the event bus.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger.Debug("Setting rate limit", "duration", duration)
	return c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: duration,
		Parallelism: 1,
	})
}

// SetMaxDepth sets the maximum crawl depth.
func (c *Crawler) SetMaxDepth(depth int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.collector != nil {
		c.collector.MaxDepth = depth
		c.Logger.Debug("Set maximum crawl depth", "depth", depth)
	}
}

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.collector = collector
	c.Logger.Debug("Collector has been set")
}

// GetIndexManager returns the index manager interface.
func (c *Crawler) GetIndexManager() api.IndexManager {
	return c.indexManager
}

// GetMetrics returns the current crawler metrics.
func (c *Crawler) GetMetrics() *Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

// handleArticle processes discovered article content.
func (c *Crawler) handleArticle(e *colly.HTMLElement) {
	c.mu.Lock()
	c.metrics.ArticlesFound++
	c.mu.Unlock()

	content := &events.Content{
		URL:  e.Request.URL.String(),
		Type: events.TypeArticle,
	}

	// Extract content based on common selectors
	content.Title = e.ChildText("h1")
	content.Description = e.ChildText("meta[name=description]")
	content.RawContent = e.Text

	// Add metadata
	content.Metadata = map[string]string{
		"language":   e.Request.Headers.Get("Accept-Language"),
		"discovered": time.Now().UTC().Format(time.RFC3339),
		"source_url": c.baseURL,
	}

	if err := c.bus.Publish(c.ctx, content); err != nil {
		c.Logger.Error("Failed to publish content to bus", "url", content.URL, "error", err)
	}
}

// handleLink processes discovered links.
func (c *Crawler) handleLink(e *colly.HTMLElement) {
	link := e.Attr("href")
	c.Logger.Debug("Found link", "url", link)
	if err := e.Request.Visit(link); err != nil {
		c.Logger.Debug("Failed to visit link", "url", link, "error", err)
	} else {
		c.Logger.Debug("Successfully queued link for visit", "url", link)
	}
}

// handleError processes collector errors.
func (c *Crawler) handleError(r *colly.Response, err error) {
	c.mu.Lock()
	c.metrics.Errors++
	c.mu.Unlock()

	c.Logger.Error("Crawler encountered an error",
		"url", r.Request.URL.String(),
		"status_code", r.StatusCode,
		"error", err,
	)
}

// setupCallbacks sets the callbacks for the collector.
func (c *Crawler) setupCallbacks() {
	c.collector.OnHTML("article", c.handleArticle)
	c.collector.OnHTML("a[href]", c.handleLink)
	c.collector.OnError(c.handleError)
	c.collector.OnScraped(func(r *colly.Response) {
		c.mu.Lock()
		c.metrics.PagesVisited++
		c.metrics.LastRequestTime = time.Now()
		c.mu.Unlock()
	})
	c.Logger.Debug("Crawler callbacks set up")
}

// monitorContext cancels operations if the context ends prematurely.
func (c *Crawler) monitorContext(ctx context.Context) {
	<-ctx.Done()
	c.Logger.Warn("Crawl cancelled due to context termination", "reason", ctx.Err())
	if c.cancel != nil {
		c.cancel()
	}
}
