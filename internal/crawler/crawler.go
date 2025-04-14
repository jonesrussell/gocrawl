// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/transport"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
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
	indexManager     interfaces.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	pageProcessor    common.Processor
	state            *State
	done             chan struct{}
	wg               sync.WaitGroup
	articleChannel   chan *models.Article
	processors       []common.Processor
	linkHandler      *LinkHandler
	htmlProcessor    *HTMLProcessor
	cfg              *crawlerconfig.Config
	abortChan        chan struct{} // Channel to signal abort
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// Core Crawler Methods
// -------------------

// ValidateSource validates a source configuration.
func (c *Crawler) validateSource(ctx context.Context, sourceName string) (*types.Source, error) {
	// Get all sources
	sourceConfigs, err := c.sources.GetSources()
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}

	// If no sources are configured, return an error
	if len(sourceConfigs) == 0 {
		return nil, errors.New("no sources configured")
	}

	// Find the requested source
	var selectedSource *sourceutils.SourceConfig
	for i := range sourceConfigs {
		if sourceConfigs[i].Name == sourceName {
			selectedSource = &sourceConfigs[i]
			break
		}
	}

	// If source not found, return an error
	if selectedSource == nil {
		return nil, fmt.Errorf("source not found: %s", sourceName)
	}

	// Convert to types.Source
	source := sourceutils.ConvertToConfigSource(selectedSource)

	// Ensure article index exists if specified
	if selectedSource.ArticleIndex != "" {
		if indexErr := c.indexManager.EnsureArticleIndex(ctx, selectedSource.ArticleIndex); indexErr != nil {
			return nil, fmt.Errorf("failed to ensure article index exists: %w", indexErr)
		}
	}

	// Ensure content index exists if specified
	if selectedSource.Index != "" {
		if contentErr := c.indexManager.EnsureContentIndex(ctx, selectedSource.Index); contentErr != nil {
			return nil, fmt.Errorf("failed to ensure content index exists: %w", contentErr)
		}
	}

	return source, nil
}

// setupCollector configures the collector with the given source settings
func (c *Crawler) setupCollector(source *types.Source) error {
	c.logger.Debug("Setting up collector",
		"max_depth", source.MaxDepth,
		"allowed_domains", source.AllowedDomains)

	opts := []colly.CollectorOption{
		colly.MaxDepth(source.MaxDepth),
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
			"default", crawlerconfig.DefaultRateLimit,
			"error", err)
		rateLimit = crawlerconfig.DefaultRateLimit
	}

	err = c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       rateLimit,
		RandomDelay: rateLimit / RandomDelayDivisor,
		Parallelism: crawlerconfig.DefaultParallelism,
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
		"parallelism", crawlerconfig.DefaultParallelism)

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
	defer close(c.abortChan) // Ensure channel is closed when done

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

	// Validate source
	source, err := c.validateSource(ctx, sourceName)
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

	// Wait for the crawler to finish
	c.collector.Wait()

	// Wait for cleanup goroutine
	select {
	case <-cleanupDone:
	case <-ctx.Done():
		return ctx.Err()
	}

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
		return c.pageProcessor
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

// SetPageProcessor sets the page processor.
func (c *Crawler) SetPageProcessor(processor common.Processor) {
	c.pageProcessor = processor
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
