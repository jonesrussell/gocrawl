// Package crawler provides the core crawling functionality for the application.
// It manages the crawling process, coordinates between components, and handles
// configuration and error management.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
)

// Constants for configuration
const (
	// TimeoutDuration is the default timeout for operations
	TimeoutDuration = 5 * time.Second
	// DefaultRateLimit is the default rate limit for requests
	DefaultRateLimit = time.Second
)

// Crawler represents a web crawler instance that manages the crawling process.
type Crawler struct {
	// Collector manages the actual web page collection
	collector *colly.Collector
	// Logger provides structured logging capabilities
	Logger common.Logger
	// Debugger handles debugging operations
	Debugger debug.Debugger
	// bus handles event publishing
	bus *events.Bus
	// baseURL is the starting point for crawling
	baseURL string
	// ctx is the main context for the crawler
	ctx context.Context
	// indexManager manages Elasticsearch indices
	indexManager api.IndexManager
}

// Ensure Crawler implements the Interface
var _ Interface = (*Crawler)(nil)

// Start begins the crawling process at the specified base URL.
func (c *Crawler) Start(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		return errors.New("base URL cannot be empty")
	}
	c.Logger.Debug("Starting crawl at base URL", "url", baseURL)
	c.baseURL = baseURL
	c.ctx = ctx

	// Ensure collector is set
	if c.collector == nil {
		return errors.New("collector not initialized")
	}

	// Set up callbacks
	c.collector.OnHTML("article", c.handleArticle)
	c.collector.OnHTML("a[href]", c.handleLink)
	c.collector.OnError(c.handleError)

	// Start crawling
	return c.collector.Visit(baseURL)
}

// Stop gracefully stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	c.Logger.Info("Stopping crawler")

	// Wait for collector to finish
	c.Wait()
	c.Logger.Info("Crawler stopped successfully")
	return nil
}

// Wait blocks until the crawler has finished processing all queued requests.
func (c *Crawler) Wait() {
	if c.collector != nil {
		c.collector.Wait()
	}
}

// Subscribe adds a content handler.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration string) error {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid rate limit duration: %w", err)
	}
	c.collector.SetRequestTimeout(d)
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

// handleArticle processes discovered article content.
func (c *Crawler) handleArticle(e *colly.HTMLElement) {
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
		c.Logger.Error("Failed to publish content", "error", err, "url", content.URL)
	}
}

// handleLink processes discovered links.
func (c *Crawler) handleLink(e *colly.HTMLElement) {
	link := e.Attr("href")
	if err := e.Request.Visit(link); err != nil {
		c.Logger.Debug("Failed to visit link", "error", err, "url", link)
	}
}

// handleError processes crawler errors.
func (c *Crawler) handleError(r *colly.Response, err error) {
	c.Logger.Error("Crawler error",
		"url", r.Request.URL.String(),
		"status_code", r.StatusCode,
		"error", err,
	)
}
