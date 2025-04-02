// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

const (
	// DefaultRandomDelayFactor is used to calculate random delay for rate limiting
	DefaultRandomDelayFactor = 2
	// DefaultParallelism is the default number of parallel requests
	DefaultParallelism = 2
)

// Crawler implements the crawler interface
type Crawler struct {
	collector        *colly.Collector
	Logger           common.Logger
	Debugger         debug.Debugger
	bus              *events.Bus
	indexManager     api.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	contentProcessor common.Processor
	testServerURL    string
}

var _ Interface = (*Crawler)(nil)

// Start starts the crawler for the given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	// Get source configuration
	source, err := c.sources.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("error getting source: %w", err)
	}

	c.Logger.Info("Starting crawler", "source", sourceName, "url", source.URL)

	// Parse the source URL to get the domain
	sourceURL, err := url.Parse(source.URL)
	if err != nil {
		return fmt.Errorf("error parsing source URL: %w", err)
	}

	// Set allowed domains only if not already configured (respect test configuration)
	// Extract host without port
	host := sourceURL.Host
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}
	c.collector.AllowedDomains = []string{host}
	c.Logger.Info("Set allowed domain", "domain", host)

	// Set up rate limiting - limit to 1 request per rate limit duration
	if rateErr := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       source.RateLimit,
		RandomDelay: 0,
		Parallelism: DefaultParallelism,
	}); rateErr != nil {
		return fmt.Errorf("error setting rate limit: %w", rateErr)
	}
	c.Logger.Info("Set rate limit", "delay", source.RateLimit, "parallelism", DefaultParallelism)

	// Set max depth
	c.SetMaxDepth(source.MaxDepth)
	c.Logger.Info("Set max depth", "depth", source.MaxDepth)

	// Configure collector
	c.collector.DetectCharset = true
	c.collector.CheckHead = true
	// Don't override domain settings if they were pre-configured
	if c.collector.DisallowedDomains == nil {
		c.collector.DisallowedDomains = nil
	}
	// Don't override URL revisit setting if it was pre-configured
	if !c.collector.AllowURLRevisit {
		c.collector.AllowURLRevisit = false
	}
	c.collector.MaxDepth = source.MaxDepth
	c.collector.Async = true

	// Start crawling
	if crawlErr := c.collector.Visit(source.URL); crawlErr != nil {
		return fmt.Errorf("error starting crawl: %w", crawlErr)
	}

	// Start processing in background
	go func() {
		c.collector.Wait()
		c.Logger.Info("Crawler finished processing", "source", sourceName)
	}()

	return nil
}

// Stop gracefully stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
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
func (c *Crawler) GetMetrics() *common.Metrics {
	// Aggregate metrics from both processors
	articleMetrics := c.articleProcessor.GetMetrics()
	contentMetrics := c.contentProcessor.GetMetrics()

	return &common.Metrics{
		ProcessedCount:     articleMetrics.ProcessedCount + contentMetrics.ProcessedCount,
		ErrorCount:         articleMetrics.ErrorCount + contentMetrics.ErrorCount,
		LastProcessedTime:  articleMetrics.LastProcessedTime,
		ProcessingDuration: articleMetrics.ProcessingDuration + contentMetrics.ProcessingDuration,
	}
}

// ProcessHTML processes HTML content from a source.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Log the element being processed
	c.Logger.Info("Processing HTML element", "tag", e.Name, "url", e.Request.URL.String())

	// Process article content
	if articleErr := c.articleProcessor.ProcessHTML(e); articleErr != nil {
		c.Logger.Error("Error processing article", "error", articleErr)
	}

	// Process general content
	if contentErr := c.contentProcessor.ProcessHTML(e); contentErr != nil {
		c.Logger.Error("Error processing content", "error", contentErr)
	}
}

// SetTestServerURL sets the test server URL for testing purposes
func (c *Crawler) SetTestServerURL(url string) {
	c.testServerURL = url
}
