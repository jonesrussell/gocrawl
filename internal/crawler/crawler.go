// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"crypto/tls"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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
	// DefaultMaxRetries is the default number of retries for failed requests
	DefaultMaxRetries = 3
	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 2
	// DefaultRateLimit is the default rate limit for requests
	DefaultRateLimit = 2 * time.Second
	// DefaultRandomDelay is the default random delay between requests
	DefaultRandomDelay = 5 * time.Second
	// DefaultBufferSize is the default size for channel buffers
	DefaultBufferSize = 100
	// DefaultMaxConcurrency is the default maximum number of concurrent requests
	DefaultMaxConcurrency = 2
	// DefaultTestSleepDuration is the default sleep duration for tests
	DefaultTestSleepDuration = 100 * time.Millisecond
	// DefaultZapFieldsCapacity is the default capacity for zap fields slice.
	DefaultZapFieldsCapacity = 2
	// CollectorStartTimeout is the timeout for collector initialization
	CollectorStartTimeout = 5 * time.Second
)

// Crawler implements the Processor interface for web crawling.
type Crawler struct {
	logger           logger.Interface
	collector        *colly.Collector
	bus              *events.EventBus
	indexManager     interfaces.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	contentProcessor common.Processor
	state            *State
	done             chan struct{}
	wg               sync.WaitGroup
	articleChannel   chan *models.Article
	processors       []common.Processor
	linkHandler      *LinkHandler
	htmlProcessor    *HTMLProcessor
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// Core Crawler Methods
// -------------------

// validateSource validates the source and its index
func (c *Crawler) validateSource(ctx context.Context, sourceName string) (*types.Source, error) {
	sourceConfig := c.sources.FindByName(sourceName)
	if sourceConfig == nil {
		return nil, fmt.Errorf("source not found: %s", sourceName)
	}

	source := sourceutils.ConvertToConfigSource(sourceConfig)
	exists, err := c.indexManager.IndexExists(ctx, source.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("index not found: %s", source.Index)
	}

	return source, nil
}

// setupCollector configures the collector with the given source settings
func (c *Crawler) setupCollector(source *types.Source) {
	c.collector = colly.NewCollector(
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
		colly.AllowedDomains(source.AllowedDomains...),
		colly.ParseHTTPErrorResponse(),
		colly.IgnoreRobotsTxt(),
	)

	// Configure transport to handle HTTP/2 better
	c.collector.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"http/1.1"},
		},
		DisableKeepAlives:     true,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   0,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})
}

// setupCallbacks configures the collector's callbacks
func (c *Crawler) setupCallbacks(ctx context.Context) {
	// Set up request callback
	c.collector.OnRequest(func(r *colly.Request) {
		if ctx.Err() != nil {
			r.Abort()
			return
		}
		c.logger.Debug("Visiting URL",
			"url", r.URL.String())
	})

	// Set up HTML processing
	c.collector.OnHTML("html", c.htmlProcessor.ProcessHTML)

	// Set up error handling
	c.collector.OnError(func(r *colly.Response, visitErr error) {
		c.logger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", visitErr)

		if errors.Is(visitErr, colly.ErrAlreadyVisited) || errors.Is(visitErr, colly.ErrMaxDepth) {
			return
		}

		c.IncrementError()
	})

	// Set up link following
	c.collector.OnHTML("a[href]", c.linkHandler.HandleLink)
}

// visitURLs visits the given URLs with error handling
func (c *Crawler) visitURLs(ctx context.Context, urls []string) {
	for _, url := range urls {
		select {
		case <-ctx.Done():
			c.logger.Debug("Crawl cancelled while visiting URLs")
			return
		default:
			if visitErr := c.collector.Visit(url); visitErr != nil {
				if errors.Is(visitErr, colly.ErrAlreadyVisited) || errors.Is(visitErr, colly.ErrMaxDepth) {
					continue
				}
				c.logger.Error("Failed to visit URL",
					"url", url,
					"error", visitErr)
				c.IncrementError()
			}
		}
	}
}

// Start begins the crawling process for a given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	c.logger.Debug("Starting crawler",
		"source", sourceName)

	// Validate source and index
	source, err := c.validateSource(ctx, sourceName)
	if err != nil {
		return err
	}

	// Initialize state
	c.state.Start(ctx, sourceName)

	// Setup collector and callbacks
	c.setupCollector(source)
	c.setupCallbacks(ctx)

	// Start crawling in a goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.logger.Debug("Starting crawl goroutine")

		crawlDone := make(chan struct{})
		crawlCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Start a goroutine to handle the crawl
		go func() {
			defer func() {
				select {
				case <-crawlDone:
					// Channel already closed
				default:
					close(crawlDone)
				}
			}()

			c.visitURLs(crawlCtx, source.StartURLs)

			c.logger.Debug("Waiting for collector to finish")
			c.collector.Wait()
			c.logger.Debug("Collector finished",
				"processed", c.state.GetProcessedCount(),
				"errors", c.state.GetErrorCount())
		}()

		select {
		case <-crawlDone:
			c.logger.Debug("Crawl goroutine finished")
		case <-ctx.Done():
			c.logger.Debug("Crawl context cancelled")
			cancel()
			c.collector.OnRequest(func(r *colly.Request) {
				r.Abort()
			})
			c.collector.Wait()
		}
	}()

	return nil
}

// Stop stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	c.logger.Debug("Stopping crawler")
	if !c.state.IsRunning() {
		c.logger.Debug("Crawler already stopped")
		return nil
	}

	// Cancel the context
	c.state.Cancel()
	c.logger.Debug("Context cancelled")

	// Stop accepting new requests
	c.collector.OnRequest(func(r *colly.Request) {
		r.Abort()
	})

	// Create a done channel for the wait group
	waitDone := make(chan struct{})

	// Start a goroutine to wait for the wait group
	go func() {
		c.logger.Debug("Waiting for wait group")
		c.wg.Wait()
		c.logger.Debug("Wait group finished")
		close(waitDone)
	}()

	// Wait for either the wait group to finish or the context to be done
	select {
	case <-waitDone:
		c.state.Stop()
		c.logger.Debug("Crawler stopped successfully")
		return nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			c.logger.Warn("Crawler shutdown timed out",
				"timeout", ctx.Err())
		} else {
			c.logger.Warn("Crawler shutdown cancelled",
				"error", ctx.Err())
		}
		return ctx.Err()
	}
}

// Wait waits for the crawler to finish.
func (c *Crawler) Wait() {
	c.wg.Wait()
}

// Done returns a channel that's closed when the crawler is done.
func (c *Crawler) Done() <-chan struct{} {
	return c.done
}

// IsRunning returns whether the crawler is running.
func (c *Crawler) IsRunning() bool {
	return c.state.IsRunning()
}

// Context returns the crawler's context.
func (c *Crawler) Context() context.Context {
	return c.state.Context()
}

// Cancel cancels the crawler's context.
func (c *Crawler) Cancel() {
	c.state.Cancel()
}

// State Management
// ---------------

// CurrentSource returns the current source being crawled.
func (c *Crawler) CurrentSource() string {
	return c.state.CurrentSource()
}

// Metrics Management
// -----------------

// GetMetrics returns the crawler metrics.
func (c *Crawler) GetMetrics() *common.Metrics {
	return &common.Metrics{
		ProcessedCount:     c.state.GetProcessedCount(),
		ErrorCount:         c.state.GetErrorCount(),
		LastProcessedTime:  c.state.GetLastProcessedTime(),
		ProcessingDuration: c.state.GetProcessingDuration(),
	}
}

// IncrementProcessed increments the processed count.
func (c *Crawler) IncrementProcessed() {
	c.state.IncrementProcessed()
}

// IncrementError increments the error count.
func (c *Crawler) IncrementError() {
	c.state.IncrementError()
}

// GetProcessedCount returns the number of processed items.
func (c *Crawler) GetProcessedCount() int64 {
	return c.state.GetProcessedCount()
}

// GetErrorCount returns the number of errors.
func (c *Crawler) GetErrorCount() int64 {
	return c.state.GetErrorCount()
}

// GetLastProcessedTime returns the time of the last processed item.
func (c *Crawler) GetLastProcessedTime() time.Time {
	return c.state.GetLastProcessedTime()
}

// GetProcessingDuration returns the total processing duration.
func (c *Crawler) GetProcessingDuration() time.Duration {
	return c.state.GetProcessingDuration()
}

// GetStartTime returns when tracking started.
func (c *Crawler) GetStartTime() time.Time {
	return c.state.GetStartTime()
}

// Update updates the metrics with new values.
func (c *Crawler) Update(startTime time.Time, processed, errors int64) {
	c.state.Update(startTime, processed, errors)
}

// Reset resets all metrics to zero.
func (c *Crawler) Reset() {
	c.state.Reset()
}

// Collector Management
// ------------------

// SetMaxDepth sets the maximum depth for the crawler.
func (c *Crawler) SetMaxDepth(depth int) {
	if c.collector == nil {
		return
	}

	config := NewCollectorConfig()
	config.MaxDepth = depth

	if err := config.Validate(); err != nil {
		c.logger.Error("Invalid max depth",
			"error", err,
			"depth", depth)
		return
	}

	c.collector.MaxDepth = config.MaxDepth
}

// SetRateLimit sets the rate limit for the crawler.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if c.collector == nil {
		return errors.New("collector is nil")
	}

	config := NewCollectorConfig()
	config.RateLimit = duration

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       config.RateLimit,
		RandomDelay: 0,
		Parallelism: config.MaxConcurrency,
	})
	if err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	return nil
}

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.collector = collector
}

// Processor Management
// ------------------

// selectProcessor selects the appropriate processor for the given HTML element
func (c *Crawler) selectProcessor(e *colly.HTMLElement) common.Processor {
	contentType := c.htmlProcessor.detectContentType(e)

	// Try to get a processor for the specific content type
	processor := c.getProcessorForType(contentType)
	if processor != nil {
		return processor
	}

	// Fallback: Try additional processors
	for _, p := range c.processors {
		if p.CanProcess(e) {
			return p
		}
	}

	return nil
}

// getProcessorForType returns a processor for the given content type
func (c *Crawler) getProcessorForType(contentType common.ContentType) common.Processor {
	switch contentType {
	case common.ContentTypeArticle:
		return c.articleProcessor
	case common.ContentTypePage:
		return c.contentProcessor
	case common.ContentTypeVideo, common.ContentTypeImage, common.ContentTypeHTML, common.ContentTypeJob:
		// Try to find a processor for the specific content type
		for _, p := range c.processors {
			if p.ContentType() == contentType {
				return p
			}
		}
	}
	return nil
}

// AddProcessor adds a new processor to the crawler.
func (c *Crawler) AddProcessor(processor common.Processor) {
	c.processors = append(c.processors, processor)
}

// SetArticleProcessor sets the article processor.
func (c *Crawler) SetArticleProcessor(processor common.Processor) {
	c.articleProcessor = processor
}

// SetContentProcessor sets the content processor.
func (c *Crawler) SetContentProcessor(processor common.Processor) {
	c.contentProcessor = processor
}

// Event Management
// ---------------

// Subscribe subscribes to crawler events.
func (c *Crawler) Subscribe(handler events.EventHandler) {
	c.bus.Subscribe(handler)
}

// Getter Methods
// -------------

// GetLogger returns the logger.
func (c *Crawler) GetLogger() logger.Interface {
	return c.logger
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

// GetIndexManager returns the index manager.
func (c *Crawler) GetIndexManager() interfaces.IndexManager {
	return c.indexManager
}

// ProcessHTML processes the HTML content.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Detect content type and get appropriate processor
	processor := c.selectProcessor(e)
	if processor == nil {
		c.logger.Debug("No processor found for content",
			"url", e.Request.URL.String(),
			"type", c.htmlProcessor.detectContentType(e))
		c.state.IncrementProcessed()
		return
	}

	// Process the content
	err := processor.Process(c.state.Context(), e)
	if err != nil {
		c.logger.Error("Failed to process content",
			"error", err,
			"url", e.Request.URL.String(),
			"type", c.htmlProcessor.detectContentType(e))
		c.state.IncrementError()
	} else {
		c.logger.Debug("Successfully processed content",
			"url", e.Request.URL.String(),
			"type", c.htmlProcessor.detectContentType(e))
	}

	c.state.IncrementProcessed()
}
