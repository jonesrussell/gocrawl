// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

const (
	// DefaultRandomDelayFactor is used to calculate random delay for rate limiting
	DefaultRandomDelayFactor = 2
	// DefaultParallelism is the default number of parallel requests
	DefaultParallelism = 2
	// DefaultStartTimeout is the default timeout for starting the crawler
	DefaultStartTimeout = 30 * time.Second
	// DefaultStopTimeout is the default timeout for stopping the crawler
	DefaultStopTimeout = 30 * time.Second
	// DefaultPollInterval is the default interval for polling crawler status
	DefaultPollInterval = 100 * time.Millisecond
)

var (
	// ErrCrawlerTimeout is returned when the crawler times out while starting
	ErrCrawlerTimeout = errors.New("timeout starting crawler")
	// ErrCrawlerContextCancelled is returned when the context is cancelled while starting the crawler
	ErrCrawlerContextCancelled = errors.New("context cancelled while starting crawler")
)

// Crawler implements the crawler interface
type Crawler struct {
	collector        *colly.Collector
	Logger           logger.Interface
	bus              *events.Bus
	indexManager     interfaces.IndexManager
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
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	articleChannel   chan *models.Article
	processors       []common.Processor
}

var _ Interface = (*Crawler)(nil)

// setupCallbacks sets up the collector callbacks
func (c *Crawler) setupCallbacks() {
	// Let Colly handle link discovery with context awareness
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-c.ctx.Done():
			e.Request.Abort()
			return
		default:
			if visitErr := e.Request.Visit(e.Attr("href")); visitErr != nil {
				// Log expected cases as debug instead of error
				if strings.Contains(visitErr.Error(), "URL already visited") {
					c.Logger.Debug("URL already visited", "url", e.Attr("href"))
				} else if strings.Contains(visitErr.Error(), "Forbidden domain") {
					c.Logger.Debug("Skipping forbidden domain", "url", e.Attr("href"))
				} else {
					c.Logger.Error("Failed to visit link", "url", e.Attr("href"), "error", visitErr)
				}
			}
		}
	})

	// Add context-aware request handling
	c.collector.OnRequest(func(r *colly.Request) {
		select {
		case <-c.ctx.Done():
			r.Abort()
			return
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

	// Add context-aware HTML processing
	c.collector.OnHTML("*", func(e *colly.HTMLElement) {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.ProcessHTML(e)
		}
	})
}

// Start starts the crawler for the given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	// Get source configuration
	source := c.sources.FindByName(sourceName)
	if source == nil {
		return fmt.Errorf("source not found: %s", sourceName)
	}

	// Validate that required index exists
	exists, err := c.indexManager.IndexExists(ctx, source.Index)
	if err != nil {
		return fmt.Errorf("failed to check index: %w", err)
	}
	if !exists {
		return fmt.Errorf("index %s does not exist", source.Index)
	}

	// Get list of sources to validate configuration
	_, err = c.sources.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	c.Logger.Info("Starting crawler", "source", sourceName, "url", source.URL)

	// Initialize collector if not already set
	if c.collector == nil {
		// Set allowed domains from source configuration
		allowedDomains := source.AllowedDomains
		if len(allowedDomains) == 0 {
			// If no allowed domains specified, extract domain from source URL
			domain, err := sourceutils.ExtractDomain(source.URL)
			if err != nil {
				return fmt.Errorf("failed to extract domain from source URL: %w", err)
			}
			allowedDomains = []string{domain}
		}

		c.collector = colly.NewCollector(
			colly.MaxDepth(source.MaxDepth),
			colly.Async(true),
			colly.AllowedDomains(allowedDomains...),
			colly.ParseHTTPErrorResponse(),
		)
		// Disable URL revisiting
		c.collector.AllowURLRevisit = false
		// Configure collector
		c.collector.DetectCharset = true
		c.collector.CheckHead = true
		// Set user agent to avoid being blocked
		c.collector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	}

	// Set up callbacks first to ensure they're ready
	c.setupCallbacks()

	// Create a cancellable context for the crawler
	c.ctx, c.cancel = context.WithCancel(ctx)
	defer c.cancel()

	// Start the crawl by visiting the source URL
	if err := c.collector.Visit(source.URL); err != nil {
		return fmt.Errorf("failed to start crawling: %w", err)
	}

	// Wait for the collector to finish or context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.collector.Wait()
	}()

	select {
	case <-done:
		// Crawler finished normally
		return nil
	case <-c.ctx.Done():
		// Context was cancelled, abort all pending requests
		c.Logger.Info("Context cancelled, aborting crawler")
		// The context cancellation will trigger request aborts in the callbacks
		return ErrCrawlerContextCancelled
	}
}

// Stop stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	c.Logger.Info("Stopping crawler")
	c.isRunning = false

	// Cancel the crawler's context first to stop all goroutines
	if c.cancel != nil {
		c.cancel()
	}

	// Create a timeout context for stopping
	stopCtx, stopCancel := context.WithTimeout(ctx, DefaultStopTimeout)
	defer stopCancel()

	// Stop the collector with timeout
	collectorDone := make(chan struct{})
	go func() {
		defer close(collectorDone)
		c.collector.Wait()
	}()

	// Wait for collector to stop or timeout
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

	// Clean up resources
	c.collector = nil
	c.done = nil
	c.ctx = nil
	c.cancel = nil

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

// GetIndexManager returns the index manager
func (c *Crawler) GetIndexManager() interfaces.IndexManager {
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

// GetLogger returns the logger.
func (c *Crawler) GetLogger() logger.Interface {
	return c.Logger
}

// GetSource returns the source.
func (c *Crawler) GetSource() sources.Interface {
	return c.sources
}

// GetProcessors returns the processors.
func (c *Crawler) GetProcessors() []common.Processor {
	return c.processors
}

// GetArticleChannel returns the article channel.
func (c *Crawler) GetArticleChannel() chan *models.Article {
	return c.articleChannel
}
