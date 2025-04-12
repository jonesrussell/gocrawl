// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// Core Crawler Methods
// -------------------

// Start begins the crawling process for a given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	c.logger.Debug("Starting crawler",
		"source", sourceName)

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
		c.logger.Debug("Starting crawl goroutine")

		// Create a done channel for the crawl
		crawlDone := make(chan struct{})
		defer close(crawlDone)

		// Start a goroutine to handle the crawl
		go func() {
			defer close(crawlDone)
			if crawlErr := c.crawl(source); crawlErr != nil {
				c.logger.Error("Failed to crawl source",
					"source", source.Name,
					"error", crawlErr)
			}
		}()

		// Wait for either the crawl to complete or the context to be cancelled
		select {
		case <-crawlDone:
			c.logger.Debug("Crawl goroutine finished")
		case <-ctx.Done():
			c.logger.Debug("Crawl context cancelled")
			// Ensure the collector is stopped
			if c.collector != nil {
				c.collector.Wait()
			}
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

	// Create a done channel for the wait group
	waitDone := make(chan struct{})
	defer close(waitDone)

	// Start a goroutine to wait for the wait group
	go func() {
		defer close(waitDone)
		c.logger.Debug("Waiting for wait group")
		c.wg.Wait()
		c.logger.Debug("Wait group finished")
	}()

	// Create a context with timeout for collector initialization
	collectorCtx, collectorCancel := context.WithTimeout(ctx, CollectorStartTimeout)
	defer collectorCancel()

	// Stop the collector
	if c.collector != nil {
		c.logger.Debug("Stopping collector")
		// Use collectorCtx to ensure we don't wait indefinitely
		select {
		case <-collectorCtx.Done():
			c.logger.Warn("Collector stop timed out",
				"timeout", collectorCtx.Err())
		default:
			c.collector.Wait()
		}
	}

	// Wait for either the wait group to finish or the context to be done
	select {
	case <-waitDone:
		c.state.Stop()
		c.logger.Debug("Crawler stopped successfully")
		return nil
	case <-ctx.Done():
		// If we hit the deadline, log the timeout and return
		c.logger.Warn("Crawler shutdown timed out",
			"timeout", ctx.Err())
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

// configureCollector configures the collector with the given source configuration.
func (c *Crawler) configureCollector(source *sourceutils.SourceConfig) {
	if source == nil {
		return
	}

	// Create collector config
	config := NewCollectorConfig()
	if source.RateLimit > 0 {
		config.RateLimit = source.RateLimit
	}
	if source.MaxDepth > 0 {
		config.MaxDepth = source.MaxDepth
	}

	// Set rate limit
	if err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: config.RateLimit,
		Parallelism: config.MaxConcurrency,
	}); err != nil {
		c.logger.Error("Failed to set rate limit",
			"error", err,
			"rateLimit", config.RateLimit,
			"parallelism", config.MaxConcurrency)
	}

	// Set max depth
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

// SetCollector sets the collector for the crawler.
func (c *Crawler) SetCollector(collector *colly.Collector) {
	c.collector = collector
}

// Processor Management
// ------------------

// selectProcessor selects the appropriate processor for the given HTML element
func (c *Crawler) selectProcessor(e *colly.HTMLElement) common.Processor {
	contentType := c.detectContentType(e)

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

// detectContentType detects the type of content in the HTML element
func (c *Crawler) detectContentType(e *colly.HTMLElement) common.ContentType {
	// Check for article-specific elements and metadata
	hasArticleTag := e.DOM.Find("article").Length() > 0
	hasArticleClass := e.DOM.Find(".article").Length() > 0
	hasArticleMeta := e.DOM.Find("meta[property='og:type'][content='article']").Length() > 0
	hasPublicationDate := e.DOM.Find("time[datetime], .published-date, .post-date").Length() > 0
	hasAuthor := e.DOM.Find(".author, .byline, meta[name='author']").Length() > 0

	// If it has multiple article indicators, it's likely an article
	if (hasArticleTag || hasArticleClass) && (hasPublicationDate || hasAuthor || hasArticleMeta) {
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

	// Default to page content type
	return common.ContentTypePage
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

// handleLink processes a single link from an HTML element.
func (c *Crawler) handleLink(e *colly.HTMLElement, source *sourceutils.SourceConfig) {
	link := e.Attr("href")
	if link == "" {
		return
	}

	// Convert relative URLs to absolute
	link = e.Request.AbsoluteURL(link)

	// Parse the URL to get the domain
	parsedURL, err := url.Parse(link)
	if err != nil {
		c.logger.Error("Failed to parse URL",
			"link", link,
			"error", err)
		return
	}

	// Check if the domain is allowed
	if !c.isDomainAllowed(parsedURL.Host, source.AllowedDomains) {
		return
	}

	// Visit the link
	if visitErr := e.Request.Visit(link); visitErr != nil {
		// Check for expected errors
		if errors.Is(visitErr, colly.ErrAlreadyVisited) || errors.Is(visitErr, colly.ErrMaxDepth) {
			return
		}

		c.logger.Error("Failed to visit link",
			"url", link,
			"error", visitErr)
	}
}

// isDomainAllowed checks if a domain is in the allowed domains list.
func (c *Crawler) isDomainAllowed(host string, allowedDomains []string) bool {
	for _, domain := range allowedDomains {
		if strings.HasSuffix(host, domain) {
			return true
		}
	}
	return false
}

// visitURLs visits a list of URLs and handles errors.
func (c *Crawler) visitURLs(urls []string) {
	for _, url := range urls {
		c.logger.Debug("Visiting URL",
			"url", url)
		if err := c.collector.Visit(url); err != nil {
			// Skip logging expected errors
			if errors.Is(err, colly.ErrAlreadyVisited) || errors.Is(err, colly.ErrMaxDepth) {
				continue
			}
			c.logger.Error("Failed to visit URL",
				"url", url,
				"error", err)
			c.IncrementError()
		}
	}
}

// crawl processes a single source.
func (c *Crawler) crawl(source *sourceutils.SourceConfig) error {
	if source == nil {
		return errors.New("source cannot be nil")
	}

	c.logger.Debug("Starting crawl for source",
		"source", source.Name,
		"urls", source.StartURLs)

	// Set up the collector for this source
	c.configureCollector(source)

	// Set up callbacks
	c.collector.OnHTML("html", c.ProcessHTML)

	// Handle errors
	c.collector.OnError(func(r *colly.Response, visitErr error) {
		c.logger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", visitErr)

		// Check for expected errors
		if errors.Is(visitErr, colly.ErrAlreadyVisited) || errors.Is(visitErr, colly.ErrMaxDepth) {
			return
		}

		// Increment error count for unexpected errors
		c.IncrementError()
	})

	// Set up link following
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		c.handleLink(e, source)
	})

	// Start visiting URLs
	c.visitURLs(source.StartURLs)

	c.logger.Debug("Waiting for collector to finish")
	// Wait for the collector to finish
	c.collector.Wait()
	c.logger.Debug("Collector finished",
		"processed", c.state.GetProcessedCount(),
		"errors", c.state.GetErrorCount())

	// Wait for all processors to finish
	c.logger.Debug("Waiting for processors to finish")
	c.wg.Wait()
	c.logger.Debug("Processors finished")

	// Signal completion
	c.logger.Debug("Signaling completion")
	close(c.done)
	c.logger.Debug("Completion signaled")

	return nil
}

// ProcessHTML processes the HTML content.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Detect content type and get appropriate processor
	processor := c.selectProcessor(e)
	if processor == nil {
		c.logger.Debug("No processor found for content",
			"url", e.Request.URL.String(),
			"type", c.detectContentType(e))
		c.state.IncrementProcessed()
		return
	}

	// Process the content
	err := processor.Process(c.state.Context(), e)
	if err != nil {
		c.logger.Error("Failed to process content",
			"error", err,
			"url", e.Request.URL.String(),
			"type", c.detectContentType(e))
		c.state.IncrementError()
	} else {
		c.logger.Debug("Successfully processed content",
			"url", e.Request.URL.String(),
			"type", c.detectContentType(e))
	}

	c.state.IncrementProcessed()
}
