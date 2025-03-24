// Package crawler provides the core crawling functionality for the application.
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
	TimeoutDuration  = 5 * time.Second // Default timeout for operations
	DefaultRateLimit = time.Second     // Default rate limit for requests
)

// Crawler represents a web crawler instance.
type Crawler struct {
	collector    *colly.Collector
	Logger       common.Logger
	Debugger     debug.Debugger
	bus          *events.Bus
	baseURL      string
	ctx          context.Context
	cancel       context.CancelFunc // To cancel the ongoing crawling process
	indexManager api.IndexManager
}

var _ Interface = (*Crawler)(nil)

// Start begins the crawling process at the specified base URL.
func (c *Crawler) Start(ctx context.Context, baseURL string) error {
	if baseURL == "" {
		return errors.New("base URL cannot be empty")
	}
	c.baseURL = baseURL
	c.ctx, c.cancel = context.WithCancel(ctx) // Use a new context with a cancel function

	// Ensure collector is initialized
	if c.collector == nil {
		return errors.New("collector not initialized")
	}

	c.Logger.Info("Starting crawl", "url", baseURL)
	c.setupCallbacks()

	// Start crawling
	err := c.collector.Visit(baseURL)
	if err != nil {
		c.Logger.Error("Error visiting base URL", "url", baseURL, "error", err)
		return fmt.Errorf("error starting crawl for base URL %s: %w", baseURL, err)
	}

	// Monitor context for cancellation while waiting
	go c.monitorContext(c.ctx)

	return nil
}

// Stop gracefully stops the crawler, respecting the provided context.
func (c *Crawler) Stop(ctx context.Context) error {
	c.Logger.Info("Stopping crawler")

	// Combine the provided context with the internal context (c.ctx)
	stopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if c.cancel != nil {
		c.cancel() // Cancel the internal crawling process
	}

	// Wait for the collector or cancellation to complete
	done := make(chan struct{})

	go func() {
		c.Wait() // Block until the collector finishes
		close(done)
	}()

	select {
	case <-stopCtx.Done(): // Respect the provided context
		c.Logger.Warn("Context for Stop was cancelled", "error", stopCtx.Err())
		return stopCtx.Err()
	case <-done: // Collector finished successfully
		c.Logger.Info("Crawler stopped successfully")
		return nil
	}
}

// Wait blocks until the crawler has finished processing all queued requests.
func (c *Crawler) Wait() {
	if c.collector != nil {
		select {
		case <-c.ctx.Done():
			c.Logger.Debug("Wait cancelled via context", "error", c.ctx.Err())
		default:
			c.collector.Wait() // Block until all requests are complete
		}
	}
}

// Subscribe adds a content handler to the event bus.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration string) error {
	d, err := time.ParseDuration(duration)
	if err != nil {
		c.Logger.Error("Invalid rate limit duration", "input", duration, "error", err)
		return fmt.Errorf("invalid rate limit duration: %w", err)
	}
	c.Logger.Debug("Setting rate limit", "duration", d)
	c.collector.SetRequestTimeout(d)
	return nil
}

// SetMaxDepth sets the maximum crawl depth.
func (c *Crawler) SetMaxDepth(depth int) {
	if c.collector != nil {
		c.collector.MaxDepth = depth
		c.Logger.Debug("Set maximum crawl depth", "depth", depth)
	}
}

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.collector = collector
	c.Logger.Debug("Collector has been set")
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
		c.Logger.Error("Failed to publish content to bus", "url", content.URL, "error", err)
	}
}

// handleLink processes discovered links.
func (c *Crawler) handleLink(e *colly.HTMLElement) {
	link := e.Attr("href")
	err := e.Request.Visit(link)
	if err != nil {
		c.Logger.Debug("Failed to visit link", "url", link, "error", err)
	}
}

// handleError processes collector errors.
func (c *Crawler) handleError(r *colly.Response, err error) {
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
