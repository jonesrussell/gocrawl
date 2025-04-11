// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fsd42/gocrawl/internal/logger"
	"github.com/fsd42/gocrawl/internal/metrics"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"go.uber.org/zap"
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

// Crawler implements the Processor interface for web crawling.
type Crawler struct {
	logger           logger.Interface
	metrics          *metrics.Metrics
	registry         *processorRegistry
	collector        *colly.Collector
	bus              *events.EventBus
	indexManager     interfaces.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	contentProcessor common.Processor
	state            *crawlerState
	done             chan struct{}
	wg               sync.WaitGroup
	articleChannel   chan *models.Article
	processors       []common.Processor
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// crawlerState tracks the crawler's state.
type crawlerState struct {
	startTime time.Time
	processed int64
	errors    int64
	mu        sync.RWMutex
	ctx       context.Context
	isRunning bool
}

// newCrawlerState creates a new crawler state.
func newCrawlerState() *crawlerState {
	return &crawlerState{
		startTime: time.Now(),
	}
}

// Update updates the state with new values.
func (s *crawlerState) Update(startTime time.Time, processed int64, errors int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startTime = startTime
	s.processed = processed
	s.errors = errors
}

// Reset resets the state to zero.
func (s *crawlerState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startTime = time.Now()
	s.processed = 0
	s.errors = 0
	s.isRunning = false
}

// GetProcessedCount returns the number of processed items.
func (s *crawlerState) GetProcessedCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.processed
}

// GetErrorCount returns the number of errors.
func (s *crawlerState) GetErrorCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errors
}

// IncrementProcessed increments the processed count.
func (s *crawlerState) IncrementProcessed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processed++
}

// IncrementError increments the error count.
func (s *crawlerState) IncrementError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
}

// Start initializes the state with a context.
func (s *crawlerState) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
	s.isRunning = true
}

// IsRunning returns whether the state is running.
func (s *crawlerState) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// Context returns the current context.
func (s *crawlerState) Context() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ctx
}

// NewCrawler creates a new crawler instance.
func NewCrawler(logger logger.Interface, metrics *metrics.Metrics) *Crawler {
	return &Crawler{
		logger:   logger,
		metrics:  metrics,
		registry: newProcessorRegistry(),
		state:    newCrawlerState(),
	}
}

// ContentType returns the type of content this processor can handle.
func (c *Crawler) ContentType() common.ContentType {
	return common.ContentTypeHTML
}

// CanProcess checks if the processor can handle the given content.
func (c *Crawler) CanProcess(content any) bool {
	_, ok := content.(string)
	return ok
}

// Process handles the content processing.
func (c *Crawler) Process(ctx context.Context, content any) error {
	html, ok := content.(string)
	if !ok {
		return fmt.Errorf("invalid content type: expected string, got %T", content)
	}

	processor, err := c.registry.GetProcessor(common.ContentTypeHTML)
	if err != nil {
		return fmt.Errorf("failed to get HTML processor: %w", err)
	}

	return processor.Process(ctx, html)
}

// ParseHTML parses HTML content from a reader.
func (c *Crawler) ParseHTML(r io.Reader) error {
	// Implementation details...
	return nil
}

// ExtractLinks extracts links from the parsed HTML.
func (c *Crawler) ExtractLinks() ([]string, error) {
	// Implementation details...
	return nil, nil
}

// ExtractContent extracts the main content from the parsed HTML.
func (c *Crawler) ExtractContent() (string, error) {
	// Implementation details...
	return "", nil
}

// ProcessJob processes a job and its items.
func (c *Crawler) ProcessJob(ctx context.Context, job *common.Job) error {
	if err := c.ValidateJob(job); err != nil {
		return fmt.Errorf("invalid job: %w", err)
	}

	processor, err := c.registry.GetProcessor(common.ContentTypeJob)
	if err != nil {
		return fmt.Errorf("failed to get job processor: %w", err)
	}

	return processor.Process(ctx, job)
}

// ValidateJob validates a job before processing.
func (c *Crawler) ValidateJob(job *common.Job) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	if job.URL == "" {
		return fmt.Errorf("job URL cannot be empty")
	}
	return nil
}

// RegisterProcessor registers a new content processor.
func (c *Crawler) RegisterProcessor(processor common.ContentProcessor) {
	c.registry.RegisterProcessor(processor)
}

// GetProcessor returns a processor for the given content type.
func (c *Crawler) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	return c.registry.GetProcessor(contentType)
}

// ProcessContent processes content using the appropriate processor.
func (c *Crawler) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	processor, err := c.registry.GetProcessor(contentType)
	if err != nil {
		return fmt.Errorf("failed to get processor for type %s: %w", contentType, err)
	}

	return processor.Process(ctx, content)
}

// Start initializes the processor.
func (c *Crawler) Start(ctx context.Context, source string) error {
	c.logger.Info("Starting crawler", zap.String("source", source))
	c.state.Start(ctx)
	return nil
}

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

// setupCallbacks sets up the collector callbacks
func (c *Crawler) setupCallbacks() {
	// Let Colly handle link discovery with context awareness
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-c.state.Context().Done():
			e.Request.Abort()
			return
		default:
			// Get the URL and check if it's a valid link
			urlStr := e.Attr("href")
			if urlStr == "" || urlStr == "#" || strings.HasPrefix(urlStr, "#") {
				// Skip empty links and anchor links
				return
			}

			// Process the HTML content
			c.ProcessHTML(e)

			// Log the link being processed
			if visitErr := e.Request.Visit(urlStr); visitErr != nil {
				// Log expected cases as debug instead of error
				if strings.Contains(visitErr.Error(), "URL already visited") {
					// Skip logging already visited URLs to reduce noise
					return
				} else if strings.Contains(visitErr.Error(), "Forbidden domain") {
					c.logger.Debug("Skipping forbidden domain", "url", urlStr)
				} else if strings.Contains(visitErr.Error(), "Max depth limit reached") {
					// Check if this is actually a forbidden domain
					parsedURL, err := url.Parse(urlStr)
					if err == nil {
						domain := parsedURL.Hostname()
						isAllowed := false
						for _, allowedDomain := range c.collector.AllowedDomains {
							if strings.HasSuffix(domain, allowedDomain) {
								isAllowed = true
								break
							}
						}
						if !isAllowed {
							c.logger.Debug("Skipping forbidden domain", "url", urlStr)
						} else {
							c.logger.Debug("Max depth limit reached", "url", urlStr)
						}
					} else {
						c.logger.Debug("Max depth limit reached", "url", urlStr)
					}
				} else if strings.Contains(visitErr.Error(), "Missing URL") {
					// Skip missing URL errors as they're usually from invalid links
					return
				} else {
					c.logger.Error("Failed to visit link", "url", urlStr, "error", visitErr)
				}
			}
		}
	})

	// Add context-aware request handling
	c.collector.OnRequest(func(r *colly.Request) {
		select {
		case <-c.state.Context().Done():
			r.Abort()
			return
		default:
			c.logger.Info("Crawling", "url", r.URL.String())
		}
	})

	c.collector.OnResponse(func(r *colly.Response) {
		select {
		case <-c.state.Context().Done():
			return
		default:
			c.logger.Info("Crawled", "url", r.Request.URL.String(), "status", r.StatusCode)
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		select {
		case <-c.state.Context().Done():
			return
		default:
			c.logger.Error("Error while crawling",
				"url", r.Request.URL.String(),
				"status", r.StatusCode,
				"error", err)
		}
	})
}

// configureCollector configures the collector with the given source configuration.
func (c *Crawler) configureCollector(source *sourceutils.SourceConfig) error {
	// Set allowed domains from source configuration
	allowedDomains := source.AllowedDomains
	if len(allowedDomains) == 0 {
		// If no allowed domains specified, extract domain from source URL
		domain, err := sourceutils.ExtractDomain(source.URL)
		if err != nil {
			return WrapError(err, "failed to extract domain from source URL")
		}
		// Add both www and non-www versions of the domain
		if strings.HasPrefix(domain, "www.") {
			allowedDomains = []string{domain, strings.TrimPrefix(domain, "www.")}
		} else {
			allowedDomains = []string{domain, "www." + domain}
		}
	}

	c.logger.Debug("Setting up collector", "allowed_domains", allowedDomains, "source_url", source.URL)

	// Create collector with basic settings, ignoring global domain restrictions
	c.collector = colly.NewCollector(
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
		colly.ParseHTTPErrorResponse(),
	)

	// Set allowed domains explicitly to override any global settings
	c.collector.AllowedDomains = allowedDomains

	// Set up rate limiting
	err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       source.RateLimit,
		RandomDelay: 0,
		Parallelism: DefaultParallelism,
	})
	if err != nil {
		return WrapError(err, "failed to set rate limit")
	}

	// Disable URL revisiting
	c.collector.AllowURLRevisit = false
	// Configure collector
	c.collector.DetectCharset = true
	c.collector.CheckHead = true
	// Set user agent to avoid being blocked
	c.collector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	return nil
}

// Start starts the crawler for the given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	// Get source configuration
	source := c.sources.FindByName(sourceName)
	if source == nil {
		return ErrSourceNotFound
	}

	// Validate that required index exists
	exists, err := c.indexManager.IndexExists(ctx, source.Index)
	if err != nil {
		return WrapError(err, "failed to check index")
	}
	if !exists {
		return ErrIndexNotFound
	}

	// Get list of sources to validate configuration
	_, err = c.sources.GetSources()
	if err != nil {
		return WrapError(err, "failed to get sources")
	}

	c.logger.Info("Starting crawler", "source", sourceName, "url", source.URL)

	// Always create a new collector for each start
	c.collector = nil
	err = c.configureCollector(source)
	if err != nil {
		return WrapError(err, "failed to configure collector")
	}

	// Set up callbacks first to ensure they're ready
	c.setupCallbacks()

	// Initialize state
	c.state = newCrawlerState()
	c.state.Start(ctx)

	// Start the crawler
	c.done = make(chan struct{})
	c.articleChannel = make(chan *models.Article, 100)

	// Start the crawler in a goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(c.done)
		defer close(c.articleChannel)

		// Start the crawler
		err := c.collector.Visit(source.URL)
		if err != nil {
			c.logger.Error("Failed to start crawler", "error", err)
			return
		}

		// Wait for the crawler to finish
		c.collector.Wait()
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

	// Wait for the crawler to stop
	select {
	case <-c.done:
		c.state.Stop()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe subscribes to crawler events.
func (c *Crawler) Subscribe(handler events.EventHandler) {
	c.bus.Subscribe(handler)
}

// SetRateLimit sets the rate limit for the crawler.
func (c *Crawler) SetRateLimit(duration time.Duration) error {
	if c.collector == nil {
		return ErrInvalidConfig
	}

	err := c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       duration,
		RandomDelay: 0,
		Parallelism: DefaultParallelism,
	})
	if err != nil {
		return WrapError(err, "failed to set rate limit")
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

// GetMetrics returns the crawler metrics.
func (c *Crawler) GetMetrics() *common.Metrics {
	return &common.Metrics{
		ProcessedCount:     c.state.GetProcessedCount(),
		ErrorCount:         c.state.GetErrorCount(),
		LastProcessedTime:  c.state.GetLastProcessedTime(),
		ProcessingDuration: c.state.GetProcessingDuration(),
	}
}

// ProcessHTML processes the HTML content.
func (c *Crawler) ProcessHTML(e *colly.HTMLElement) {
	// Process the HTML content
	if c.contentProcessor != nil && c.contentProcessor.CanProcess(e) {
		err := c.contentProcessor.Process(c.state.Context(), e)
		if err != nil {
			c.logger.Error("Failed to process content", "error", err)
			c.state.IncrementError()
		}
	}

	// Process the article
	if c.articleProcessor != nil && c.articleProcessor.CanProcess(e) {
		err := c.articleProcessor.Process(c.state.Context(), e)
		if err != nil {
			c.logger.Error("Failed to process article", "error", err)
			c.state.IncrementError()
		}
	}

	// Process with additional processors
	for _, processor := range c.processors {
		if processor.CanProcess(e) {
			err := processor.Process(c.state.Context(), e)
			if err != nil {
				c.logger.Error("Failed to process with additional processor",
					"processor", processor.ContentType(),
					"error", err)
				c.state.IncrementError()
			}
		}
	}

	c.state.IncrementProcessed()
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
func (c *Crawler) Update(startTime time.Time, processed int64, errors int64) {
	c.state.Update(startTime, processed, errors)
}

// Reset resets all metrics to zero.
func (c *Crawler) Reset() {
	c.state.Reset()
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
