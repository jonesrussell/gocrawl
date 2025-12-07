// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	colly "github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common/transport"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

const (
	// RandomDelayDivisor is used to calculate random delay from rate limit
	RandomDelayDivisor = 2
)

// Crawler implements the Processor interface for web crawling.
type Crawler struct {
	logger           logger.Interface
	collector        *colly.Collector
	bus              *events.EventBus
	indexManager     storagetypes.IndexManager
	sources          sources.Interface
	articleProcessor content.Processor
	pageProcessor    content.Processor
	state            *State
	done             chan struct{}
	wg               sync.WaitGroup
	articleChannel   chan *models.Article
	processors       []content.Processor
	linkHandler      *LinkHandler
	htmlProcessor    *HTMLProcessor
	cfg              *crawler.Config
	abortChan        chan struct{} // Channel to signal abort
	maxDepthOverride int           // Override for source's max_depth (0 means use source default)
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// Core Crawler Methods
// -------------------

// setupCollector configures the collector with the given source settings
func (c *Crawler) setupCollector(source *configtypes.Source) error {
	// Use override if set, otherwise use source's max depth
	maxDepth := source.MaxDepth
	if c.maxDepthOverride > 0 {
		maxDepth = c.maxDepthOverride
		c.logger.Info("Using max_depth override", "override", maxDepth, "source_default", source.MaxDepth)
	}

	c.logger.Debug("Setting up collector",
		"max_depth", maxDepth,
		"allowed_domains", source.AllowedDomains)

	opts := []colly.CollectorOption{
		colly.MaxDepth(maxDepth),
		colly.Async(true),
		colly.ParseHTTPErrorResponse(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(c.cfg.UserAgent),
		colly.AllowURLRevisit(),
	}

	// Only set allowed domains if they are configured
	if len(source.AllowedDomains) > 0 {
		opts = append(opts, colly.AllowedDomains(source.AllowedDomains...))
	}

	c.collector = colly.NewCollector(opts...)

	// Parse and set rate limit
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		c.logger.Error("Failed to parse rate limit, using default",
			"rate_limit", source.RateLimit,
			"default", crawler.DefaultRateLimit,
			"error", err)
		rateLimit = crawler.DefaultRateLimit
	}

	err = c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       rateLimit,
		RandomDelay: rateLimit / RandomDelayDivisor,
		Parallelism: crawler.DefaultParallelism,
	})
	if err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Configure transport with more reasonable settings
	tlsConfig, err := transport.NewTLSConfig(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create TLS configuration: %w", err)
	}

	c.collector.WithTransport(&http.Transport{
		TLSClientConfig:       tlsConfig,
		DisableKeepAlives:     false,
		MaxIdleConns:          transport.DefaultMaxIdleConns,
		MaxIdleConnsPerHost:   transport.DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:       transport.DefaultIdleConnTimeout,
		ResponseHeaderTimeout: transport.DefaultResponseHeaderTimeout,
		ExpectContinueTimeout: transport.DefaultExpectContinueTimeout,
	})

	if c.cfg.TLS.InsecureSkipVerify {
		c.logger.Warn("TLS certificate verification is disabled. This is not recommended for production use.",
			"component", "crawler",
			"source", source.Name,
			"warning", "This makes HTTPS connections vulnerable to man-in-the-middle attacks")
	}

	c.logger.Debug("Collector configured",
		"max_depth", source.MaxDepth,
		"allowed_domains", source.AllowedDomains,
		"rate_limit", rateLimit,
		"parallelism", crawler.DefaultParallelism)

	return nil
}

// setupCallbacks configures the collector's callbacks
func (c *Crawler) setupCallbacks(ctx context.Context) {
	// Set up response callback
	c.collector.OnResponse(func(r *colly.Response) {
		c.logger.Debug("Received response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"headers", r.Headers)
	})

	// Set up request callback
	c.collector.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
			return
		case <-c.abortChan:
			r.Abort()
			return
		default:
			c.logger.Debug("Visiting URL",
				"url", r.URL.String())
		}
	})

	// Set up HTML processing
	c.collector.OnHTML("html", func(e *colly.HTMLElement) {
		c.ProcessHTML(e)
	})

	// Set up error handling
	c.collector.OnError(func(r *colly.Response, visitErr error) {
		errMsg := visitErr.Error()

		// Check if this is an expected/non-critical error (log at debug)
		isExpectedError := errors.Is(visitErr, ErrAlreadyVisited) ||
			errors.Is(visitErr, ErrMaxDepth) ||
			errors.Is(visitErr, ErrForbiddenDomain) ||
			strings.Contains(errMsg, "forbidden domain") ||
			strings.Contains(errMsg, "Forbidden domain") ||
			strings.Contains(errMsg, "max depth") ||
			strings.Contains(errMsg, "Max depth") ||
			strings.Contains(errMsg, "already visited") ||
			strings.Contains(errMsg, "Already visited") ||
			strings.Contains(errMsg, "Not following redirect")

		if isExpectedError {
			// These are expected conditions, log at debug level
			c.logger.Debug("Expected error while crawling",
				"url", r.Request.URL.String(),
				"status", r.StatusCode,
				"error", errMsg)
			return
		}

		// Check if this is a timeout (log at warn level - common but still an issue)
		isTimeout := strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "Timeout") ||
			strings.Contains(errMsg, "deadline exceeded") ||
			strings.Contains(errMsg, "context deadline exceeded")

		if isTimeout {
			// Timeouts are common when crawling, log at warn level
			c.logger.Warn("Timeout while crawling",
				"url", r.Request.URL.String(),
				"status", r.StatusCode,
				"error", errMsg)
			c.IncrementError()
			return
		}

		// Log actual errors
		c.logger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", visitErr)

		c.IncrementError()
	})

	// Set up link following
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		c.linkHandler.HandleLink(e)
	})

	// Set up scraped callback to handle abort
	c.collector.OnScraped(func(r *colly.Response) {
		select {
		case <-ctx.Done():
			r.Request.Abort()
			return
		case <-c.abortChan:
			r.Request.Abort()
			return
		default:
			// Continue processing
		}
	})
}

// Start begins the crawling process for a given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	c.logger.Debug("Starting crawler",
		"source", sourceName,
		"debug_enabled", c.cfg.Debug,
	)

	// Initialize abort channel
	c.abortChan = make(chan struct{})
	var abortChanOnce sync.Once

	// Start cleanup goroutine
	cleanupDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(c.cfg.CleanupInterval)
		defer ticker.Stop()
		defer close(cleanupDone)

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.abortChan:
				return
			case <-ticker.C:
				c.cleanupResources()
			}
		}
	}()

	// Ensure abortChan is closed on exit
	defer func() {
		abortChanOnce.Do(func() {
			close(c.abortChan)
		})
	}()

	// Validate source
	source, err := c.sources.ValidateSource(ctx, sourceName, c.indexManager)
	if err != nil {
		return fmt.Errorf("failed to validate source: %w", err)
	}

	// Set up collector
	err = c.setupCollector(source)
	if err != nil {
		return fmt.Errorf("failed to setup collector: %w", err)
	}

	// Set up callbacks
	c.setupCallbacks(ctx)

	// Start the crawler state
	c.state.Start(ctx, sourceName)

	// Visit the source URL
	if visitErr := c.collector.Visit(source.URL); visitErr != nil {
		return fmt.Errorf("failed to visit source URL: %w", visitErr)
	}

	// Wait for the crawler to finish, but respect context cancellation
	// Run Wait() in a goroutine so we can check for context cancellation
	waitDone := make(chan struct{})
	go func() {
		c.collector.Wait()
		close(waitDone)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-waitDone:
		// Collector finished normally
		c.logger.Debug("Collector finished normally")
	case <-ctx.Done():
		// Context was cancelled - abort all pending requests
		c.logger.Info("Context cancelled, aborting collector")
		// Signal abort to stop new requests (safe to call multiple times)
		abortChanOnce.Do(func() {
			close(c.abortChan)
		})
		// Wait a bit for in-flight requests to abort, then return
		select {
		case <-waitDone:
			// Collector finished after abort
		case <-time.After(2 * time.Second):
			// Timeout waiting for collector to finish
			c.logger.Warn("Collector did not finish within timeout after cancellation")
		}
		return ctx.Err()
	}

	// Signal cleanup goroutine to stop by closing abortChan
	// This will cause the cleanup goroutine to exit (safe to call multiple times)
	abortChanOnce.Do(func() {
		close(c.abortChan)
	})

	// Wait for cleanup goroutine to finish
	select {
	case <-cleanupDone:
		c.logger.Debug("Cleanup goroutine finished")
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		// Timeout after 5 seconds - cleanup goroutine should have finished by now
		c.logger.Warn("Cleanup goroutine did not finish within timeout")
	}

	// Stop the crawler state
	c.state.Stop()

	return nil
}

// cleanupResources performs periodic cleanup of crawler resources
func (c *Crawler) cleanupResources() {
	c.logger.Debug("Cleaning up crawler resources")

	// Clean up article channel
	select {
	case <-c.articleChannel: // Try to read one item
	default: // Channel is empty
	}

	// Clean up processors
	for _, p := range c.processors {
		if cleaner, ok := p.(interface{ Cleanup() }); ok {
			cleaner.Cleanup()
		}
	}

	// Clean up state
	c.state.Reset()

	c.logger.Debug("Finished cleaning up crawler resources")
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

	// Signal abort to all goroutines
	close(c.abortChan)

	// Wait for the collector to finish
	c.collector.Wait()

	// Create a done channel for the wait group
	waitDone := make(chan struct{})

	// Start a goroutine to wait for the wait group
	go func() {
		c.wg.Wait()
		close(waitDone)
	}()

	// Wait for either the wait group to finish or the context to be done
	select {
	case <-waitDone:
		c.state.Stop()
		c.cleanupResources() // Final cleanup
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

// Wait waits for the crawler to complete
func (c *Crawler) Wait() error {
	c.wg.Wait()
	return nil
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
func (c *Crawler) GetMetrics() *metrics.Metrics {
	return &metrics.Metrics{
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
// If the collector hasn't been created yet, this sets an override that will be used
// when the collector is created. Otherwise, it updates the existing collector.
func (c *Crawler) SetMaxDepth(depth int) {
	config := NewCollectorConfig()
	config.MaxDepth = depth

	if err := config.Validate(); err != nil {
		c.logger.Error("Invalid max depth",
			"error", err,
			"depth", depth)
		return
	}

	if c.collector == nil {
		// Collector not created yet, store as override to use when collector is created
		c.maxDepthOverride = config.MaxDepth
		c.logger.Debug("Set max_depth override (collector not yet created)", "max_depth", config.MaxDepth)
	} else {
		// Collector exists, update it directly
		c.collector.MaxDepth = config.MaxDepth
		c.logger.Debug("Updated collector max_depth", "max_depth", config.MaxDepth)
	}
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
func (c *Crawler) selectProcessor(e *colly.HTMLElement) content.Processor {
	contentType := c.htmlProcessor.detectContentType(e)

	// Try to get a processor for the specific content type
	processor := c.getProcessorForType(contentType)
	if processor != nil {
		return processor
	}

	// Fallback: Try additional processors
	for _, p := range c.processors {
		if p.CanProcess(contentType) {
			return p
		}
	}

	return nil
}

// getProcessorForType returns a processor for the given content type
func (c *Crawler) getProcessorForType(contentType contenttype.Type) content.Processor {
	switch contentType {
	case contenttype.Article:
		return c.articleProcessor
	case contenttype.Page:
		return c.pageProcessor
	case contenttype.Video, contenttype.Image, contenttype.HTML, contenttype.Job:
		// Try to find a processor for the specific content type
		for _, p := range c.processors {
			if p.CanProcess(contentType) {
				return p
			}
		}
	}
	return nil
}

// AddProcessor adds a new processor to the crawler.
func (c *Crawler) AddProcessor(processor content.Processor) {
	c.processors = append(c.processors, processor)
}

// SetArticleProcessor sets the article processor.
func (c *Crawler) SetArticleProcessor(processor content.Processor) {
	c.articleProcessor = processor
}

// SetPageProcessor sets the page processor.
func (c *Crawler) SetPageProcessor(processor content.Processor) {
	c.pageProcessor = processor
}

// GetProcessors returns the processors.
func (c *Crawler) GetProcessors() []content.Processor {
	processors := make([]content.Processor, 0, len(c.processors))
	for _, p := range c.processors {
		wrapper := &processorWrapper{
			processor: p,
			registry:  make([]content.ContentProcessor, 0),
		}
		processors = append(processors, wrapper)
	}
	return processors
}

// processorWrapper wraps a content.Processor to implement content.Processor
type processorWrapper struct {
	processor content.Processor
	registry  []content.ContentProcessor
}

// ContentType implements content.ContentProcessor
func (p *processorWrapper) ContentType() contenttype.Type {
	return p.processor.ContentType()
}

// CanProcess implements content.ContentProcessor
func (p *processorWrapper) CanProcess(content contenttype.Type) bool {
	return p.processor.CanProcess(content)
}

// Process implements content.ContentProcessor
func (p *processorWrapper) Process(ctx context.Context, content any) error {
	return p.processor.Process(ctx, content)
}

// ParseHTML implements content.HTMLProcessor
func (p *processorWrapper) ParseHTML(r io.Reader) error {
	return p.processor.ParseHTML(r)
}

// ExtractLinks implements content.HTMLProcessor
func (p *processorWrapper) ExtractLinks() ([]string, error) {
	return p.processor.ExtractLinks()
}

// ExtractContent implements content.HTMLProcessor
func (p *processorWrapper) ExtractContent() (string, error) {
	return p.processor.ExtractContent()
}

// RegisterProcessor implements content.ProcessorRegistry
func (p *processorWrapper) RegisterProcessor(processor content.ContentProcessor) {
	p.registry = append(p.registry, processor)
}

// GetProcessor implements content.ProcessorRegistry
func (p *processorWrapper) GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error) {
	for _, processor := range p.registry {
		if processor.CanProcess(contentType) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no processor found for content type: %s", contentType)
}

// ProcessContent implements content.ProcessorRegistry
func (p *processorWrapper) ProcessContent(ctx context.Context, contentType contenttype.Type, content any) error {
	processor, err := p.GetProcessor(contentType)
	if err != nil {
		return err
	}
	return processor.Process(ctx, content)
}

// Start implements content.Processor
func (p *processorWrapper) Start(ctx context.Context) error {
	return p.processor.Start(ctx)
}

// Stop implements content.Processor
func (p *processorWrapper) Stop(ctx context.Context) error {
	return p.processor.Stop(ctx)
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

// GetArticleChannel returns the article channel.
func (c *Crawler) GetArticleChannel() chan *models.Article {
	return c.articleChannel
}

// GetIndexManager returns the index manager.
func (c *Crawler) GetIndexManager() storagetypes.IndexManager {
	return c.indexManager
}

// ProcessHTML processes the HTML content.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Check if context is cancelled before processing
	ctx := c.state.Context()
	select {
	case <-ctx.Done():
		// Context cancelled, abort this request
		e.Request.Abort()
		return
	default:
		// Continue processing
	}

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
		// If the error is "not implemented", log at debug level since this is expected
		// until the feature is implemented
		if err.Error() == "not implemented" {
			c.logger.Debug("Content processing not implemented",
				"url", e.Request.URL.String(),
				"type", c.htmlProcessor.detectContentType(e))
		} else {
			c.logger.Error("Failed to process content",
				"error", err,
				"url", e.Request.URL.String(),
				"type", c.htmlProcessor.detectContentType(e))
			c.state.IncrementError()
		}
	} else {
		c.logger.Debug("Successfully processed content",
			"url", e.Request.URL.String(),
			"type", c.htmlProcessor.detectContentType(e))
	}

	c.state.IncrementProcessed()
}

// GetProcessor returns a processor for the given content type.
func (c *Crawler) GetProcessor(contentType contenttype.Type) (content.Processor, error) {
	for _, p := range c.processors {
		if p.CanProcess(contentType) {
			return p, nil
		}
	}

	if contentType == contenttype.Article {
		return c.articleProcessor, nil
	}

	if contentType == contenttype.Page {
		return c.pageProcessor, nil
	}

	return nil, fmt.Errorf("no processor found for content type: %s", contentType)
}
