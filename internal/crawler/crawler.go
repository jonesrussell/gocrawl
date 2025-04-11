// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
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
)

// Crawler implements the Processor interface for web crawling.
type Crawler struct {
	logger           logger.Interface
	registry         *processorRegistry
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
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// processorRegistry implements the ProcessorRegistry interface.
type processorRegistry struct {
	processors map[common.ContentType]common.ContentProcessor
	mu         sync.RWMutex
}

// newProcessorRegistry creates a new processor registry.
func newProcessorRegistry() *processorRegistry {
	return &processorRegistry{
		processors: make(map[common.ContentType]common.ContentProcessor),
	}
}

// RegisterProcessor registers a new content processor.
func (r *processorRegistry) RegisterProcessor(processor common.ContentProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[processor.ContentType()] = processor
}

// GetProcessor returns a processor for the given content type.
func (r *processorRegistry) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	processor, ok := r.processors[contentType]
	if !ok {
		return nil, fmt.Errorf("no processor registered for content type %s", contentType)
	}
	return processor, nil
}

// configureCollector configures the collector with the given source configuration.
func (c *Crawler) configureCollector(source *sourceutils.SourceConfig) {
	if source == nil {
		return
	}

	// Set rate limit
	rateLimit := common.DefaultRateLimit
	if source.RateLimit > 0 {
		rateLimit = source.RateLimit
	}
	if err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: rateLimit,
		Parallelism: common.DefaultMaxConcurrency,
	}); err != nil {
		c.logger.Error("Failed to set rate limit",
			"error", err,
			"rateLimit", rateLimit,
			"parallelism", common.DefaultMaxConcurrency)
	}

	// Set max depth
	if source.MaxDepth > 0 {
		c.collector.MaxDepth = source.MaxDepth
	} else {
		c.collector.MaxDepth = common.DefaultMaxDepth
	}
}

// crawl processes a single source.
func (c *Crawler) crawl(source *sourceutils.SourceConfig) error {
	if source == nil {
		return errors.New("source cannot be nil")
	}

	// Set up the collector for this source
	c.configureCollector(source)

	// Set up callbacks
	c.collector.OnHTML("html", c.ProcessHTML)

	// Start visiting URLs
	for _, url := range source.StartURLs {
		if err := c.collector.Visit(url); err != nil {
			c.logger.Error("Failed to visit URL",
				"url", url,
				"error", err)
			c.IncrementError()
		}
	}

	// Wait for the collector to finish
	c.collector.Wait()

	return nil
}

// Start begins the crawling process for a given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	// Validate source exists
	source := c.sources.FindByName(sourceName)
	if source == nil {
		return fmt.Errorf("source not found: %s", sourceName)
	}

	// Validate index exists
	exists, err := c.indexManager.IndexExists(ctx, source.Index)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("index not found: %s", source.Index)
	}

	// Initialize state with context and source name
	c.state.Start(ctx, sourceName)

	// Configure collector for this source
	c.configureCollector(source)

	// Start crawling in a goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if crawlErr := c.crawl(source); crawlErr != nil {
			c.logger.Error("Failed to crawl source",
				"source", source.Name,
				"error", crawlErr)
		}
	}()

	return nil
}

// Stop stops the crawler.
func (c *Crawler) Stop(ctx context.Context) error {
	if !c.state.IsRunning() {
		return nil
	}

	// Cancel the context
	c.state.Cancel()

	// Create a done channel for the collector
	collectorDone := make(chan struct{})

	// Start a goroutine to wait for the collector and wait group
	go func() {
		defer close(collectorDone)
		c.collector.Wait()
		c.wg.Wait()
	}()

	// Wait for either the collector to finish or the context to be done
	select {
	case <-collectorDone:
		c.state.Stop()
		return nil
	case <-ctx.Done():
		// If we hit the deadline, log the timeout and return
		c.logger.Warn("Crawler shutdown timed out",
			"timeout", ctx.Err())
		return ctx.Err()
	}
}

// Subscribe subscribes to crawler events.
func (c *Crawler) Subscribe(handler events.EventHandler) {
	c.bus.Subscribe(handler)
}

// GetMetrics returns the crawler metrics.
func (c *Crawler) GetMetrics() *common.Metrics {
	return &common.Metrics{
		ProcessedCount:     c.state.GetProcessedCount(),
		ErrorCount:         c.state.GetErrorCount(),
		LastProcessedTime:  c.state.GetLastProcessedTime(),
		ProcessingDuration: c.state.GetProcessingDuration(),
	}
}

// SetRateLimit sets the rate limit for the crawler.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if c.collector == nil {
		return errors.New("collector is nil")
	}

	err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       duration,
		RandomDelay: 0,
		Parallelism: common.DefaultMaxConcurrency,
	})
	if err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	return nil
}

// SetMaxDepth sets the maximum depth for the crawler.
func (c *Crawler) SetMaxDepth(depth int) {
	if c.collector != nil {
		c.collector.MaxDepth = depth
	}
}

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.collector = collector
}

// GetIndexManager returns the index manager.
func (c *Crawler) GetIndexManager() interfaces.IndexManager {
	return c.indexManager
}

// Wait waits for the crawler to finish.
func (c *Crawler) Wait() {
	c.wg.Wait()
}

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

// IsRunning returns whether the crawler is running.
func (c *Crawler) IsRunning() bool {
	return c.state.IsRunning()
}

// StartTime returns when the crawler started.
func (c *Crawler) StartTime() time.Time {
	return c.state.StartTime()
}

// CurrentSource returns the current source being crawled.
func (c *Crawler) CurrentSource() string {
	return c.state.CurrentSource()
}

// Context returns the crawler's context.
func (c *Crawler) Context() context.Context {
	return c.state.Context()
}

// Cancel cancels the crawler's context.
func (c *Crawler) Cancel() {
	c.state.Cancel()
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

// GetStartTime returns when tracking started.
func (c *Crawler) GetStartTime() time.Time {
	return c.state.StartTime()
}

// GetLastProcessedTime returns the time of the last processed item.
func (c *Crawler) GetLastProcessedTime() time.Time {
	return c.state.GetLastProcessedTime()
}

// GetProcessingDuration returns the total processing duration.
func (c *Crawler) GetProcessingDuration() time.Duration {
	return c.state.GetProcessingDuration()
}

// Update updates the metrics with new values.
func (c *Crawler) Update(startTime time.Time, processed, errors int64) {
	c.state.Update(startTime, processed, errors)
}

// Reset resets all metrics to zero.
func (c *Crawler) Reset() {
	c.state.Reset()
}

// ProcessHTML processes the HTML content.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Get the crawler's context
	ctx := c.state.Context()

	// Detect content type
	contentType := c.detectContentType(e)

	// Process with the appropriate processor
	var processor common.Processor
	switch contentType {
	case common.ContentTypeArticle:
		processor = c.articleProcessor
	case common.ContentTypePage:
		processor = c.contentProcessor
	default:
		// Try additional processors
		for _, p := range c.processors {
			if p.CanProcess(e) {
				processor = p
				break
			}
		}
	}

	if processor != nil {
		err := processor.Process(ctx, e)
		if err != nil {
			c.logger.Error("Failed to process content",
				"error", err,
				"type", contentType)
			c.state.IncrementError()
		}
	}

	c.state.IncrementProcessed()
}

// detectContentType detects the type of content in the HTML element
func (c *Crawler) detectContentType(e *colly.HTMLElement) common.ContentType {
	// Check for article-specific elements
	if e.DOM.Find("article").Length() > 0 || e.DOM.Find(".article").Length() > 0 {
		return common.ContentTypeArticle
	}

	// Check for video content
	if e.DOM.Find("video").Length() > 0 || e.DOM.Find(".video").Length() > 0 {
		return common.ContentTypeVideo
	}

	// Check for image content
	if e.DOM.Find("img").Length() > 0 || e.DOM.Find(".image").Length() > 0 {
		return common.ContentTypeImage
	}

	// Check for job listings
	if e.DOM.Find(".job-listing").Length() > 0 || e.DOM.Find(".job-posting").Length() > 0 {
		return common.ContentTypeJob
	}

	// Default to HTML content type
	return common.ContentTypeHTML
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
