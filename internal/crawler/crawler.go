// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
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
	bus              *events.Bus
	indexManager     api.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	contentProcessor common.Processor
	testServerURL    string
	processedCount   int64
	errorCount       int64
	startTime        time.Time
	isRunning        bool
	done             chan struct{}
	ctx              context.Context
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
	c.Logger.Debug("Set allowed domain", "domain", host)

	// Set up rate limiting - limit to 1 request per rate limit duration
	if rateErr := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       source.RateLimit,
		RandomDelay: 0,
		Parallelism: DefaultParallelism,
	}); rateErr != nil {
		return fmt.Errorf("error setting rate limit: %w", rateErr)
	}
	c.Logger.Debug("Set rate limit", "delay", source.RateLimit, "parallelism", DefaultParallelism)

	// Set max depth
	c.SetMaxDepth(source.MaxDepth)
	c.Logger.Debug("Set max depth", "depth", source.MaxDepth)

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

	// Reset metrics and state
	c.processedCount = 0
	c.errorCount = 0
	c.startTime = time.Now()
	c.isRunning = true
	c.done = make(chan struct{})
	c.ctx = ctx

	// Create a timeout context for starting the crawler
	startCtx, startCancel := context.WithTimeout(ctx, 30*time.Second)
	defer startCancel()

	// Start crawling with timeout
	crawlErr := make(chan error)
	go func() {
		defer close(crawlErr)
		if err := c.collector.Visit(source.URL); err != nil {
			crawlErr <- fmt.Errorf("error starting crawl: %w", err)
		}
	}()

	select {
	case err := <-crawlErr:
		if err != nil {
			return err
		}
	case <-startCtx.Done():
		return fmt.Errorf("timeout starting crawler")
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while starting crawler")
	}

	// Start processing in background
	go func() {
		defer close(c.done)
		for {
			select {
			case <-ctx.Done():
				c.isRunning = false
				c.Logger.Info("Context cancelled, stopping crawler", "source", sourceName)
				return
			case <-time.After(100 * time.Millisecond):
				if !c.isRunning {
					c.Logger.Info("Crawler finished processing", "source", sourceName)
					return
				}
			}
		}
	}()

	// Let Colly handle link discovery with context awareness
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-c.ctx.Done():
			return
		default:
			e.Request.Visit(e.Attr("href"))
		}
	})

	// Add context-aware request handling
	c.collector.OnRequest(func(r *colly.Request) {
		select {
		case <-c.ctx.Done():
			r.Abort()
		default:
			c.Logger.Debug("Visiting", "url", r.URL.String())
		}
	})

	c.collector.OnResponse(func(r *colly.Response) {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.Logger.Debug("Visited", "url", r.Request.URL.String(), "status", r.StatusCode)
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.Logger.Error("Error while crawling",
				"url", r.Request.URL.String(),
				"status", r.StatusCode,
				"error", err)
		}
	})

	return nil
}

// Stop stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	c.Logger.Info("Stopping crawler")
	c.isRunning = false

	// Create a timeout context for stopping
	stopCtx, stopCancel := context.WithTimeout(ctx, 30*time.Second)
	defer stopCancel()

	// Stop the collector with timeout
	collectorDone := make(chan struct{})
	go func() {
		defer close(collectorDone)
		c.collector.Wait()
	}()

	select {
	case <-collectorDone:
		c.Logger.Info("Collector stopped successfully")
	case <-stopCtx.Done():
		c.Logger.Warn("Timeout waiting for collector to stop")
	case <-ctx.Done():
		c.Logger.Warn("Context cancelled while stopping collector")
	}

	// Wait for background goroutine to finish
	if c.done != nil {
		select {
		case <-c.done:
			c.Logger.Info("Crawler stopped successfully")
		case <-stopCtx.Done():
			c.Logger.Warn("Timeout waiting for crawler to stop")
		case <-ctx.Done():
			c.Logger.Warn("Context cancelled while waiting for crawler to stop")
		}
	}

	return nil
}

// Subscribe adds a content handler to receive discovered content.
func (c *Crawler) Subscribe(handler events.Handler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the crawler's rate limit.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if rateErr := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       duration,
		RandomDelay: 0,
		Parallelism: DefaultParallelism,
	}); rateErr != nil {
		return fmt.Errorf("error setting rate limit: %w", rateErr)
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
	if c.done != nil {
		<-c.done
	}
}

// GetMetrics returns the current crawler metrics.
func (c *Crawler) GetMetrics() *common.Metrics {
	return &common.Metrics{
		ProcessedCount:     c.processedCount,
		ErrorCount:         c.errorCount,
		LastProcessedTime:  time.Now(),
		ProcessingDuration: time.Since(c.startTime),
	}
}

// ProcessHTML processes HTML content from a source.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Log the element being processed
	c.Logger.Debug("Processing HTML element", "tag", e.Name, "url", e.Request.URL.String())

	// Process article content
	c.handleArticle(e)

	// Process general content
	c.handleContent(e)
}

// handleArticle processes an article using the article processor
func (c *Crawler) handleArticle(e *colly.HTMLElement) {
	if err := c.articleProcessor.ProcessHTML(c.ctx, e); err != nil {
		c.Logger.Error("Failed to process article",
			"component", "crawler",
			"url", e.Request.URL.String(),
			"error", err)
		c.errorCount++
	} else {
		c.processedCount++
	}
}

// handleContent processes content using the content processor
func (c *Crawler) handleContent(e *colly.HTMLElement) {
	if err := c.contentProcessor.ProcessHTML(c.ctx, e); err != nil {
		c.Logger.Error("Failed to process content",
			"component", "crawler",
			"url", e.Request.URL.String(),
			"error", err)
		c.errorCount++
	} else {
		c.processedCount++
	}
}

// SetTestServerURL sets the test server URL for testing purposes
func (c *Crawler) SetTestServerURL(url string) {
	c.testServerURL = url
}
