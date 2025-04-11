// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
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
	config           *Config
	esClient         *elasticsearch.Client
}

var _ Interface = (*Crawler)(nil)
var _ CrawlerInterface = (*Crawler)(nil)
var _ CrawlerMetrics = (*Crawler)(nil)

// crawlerState tracks the crawler's state.
type crawlerState struct {
	startTime         time.Time
	processedCount    int64
	errorCount        int64
	lastProcessedTime time.Time
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	isRunning         bool
	currentSource     string
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
	s.processedCount = processed
	s.errorCount = errors
}

// Reset resets the state to zero.
func (s *crawlerState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startTime = time.Now()
	s.processedCount = 0
	s.errorCount = 0
	s.lastProcessedTime = time.Time{}
	s.isRunning = false
	s.currentSource = ""
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.ctx = nil
}

// GetProcessedCount returns the number of processed items.
func (s *crawlerState) GetProcessedCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.processedCount
}

// GetErrorCount returns the number of errors.
func (s *crawlerState) GetErrorCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorCount
}

// IncrementProcessed increments the processed count.
func (s *crawlerState) IncrementProcessed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processedCount++
	s.lastProcessedTime = time.Now()
}

// IncrementError increments the error count.
func (s *crawlerState) IncrementError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorCount++
}

// Start initializes the state with a context.
func (s *crawlerState) Start(ctx context.Context, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.isRunning = true
	s.currentSource = source
}

// Stop cleans up the state.
func (s *crawlerState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = false
	s.currentSource = ""
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.ctx = nil
}

// Cancel cancels the context.
func (s *crawlerState) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
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

// CurrentSource returns the current source.
func (s *crawlerState) CurrentSource() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSource
}

// GetLastProcessedTime returns the time of the last processed item.
func (s *crawlerState) GetLastProcessedTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastProcessedTime
}

// GetProcessingDuration returns the total processing duration.
func (s *crawlerState) GetProcessingDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isRunning {
		return 0
	}
	return time.Since(s.startTime)
}

// StartTime returns the start time of the crawler.
func (s *crawlerState) StartTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

// zapLoggerWrapper wraps a *zap.Logger to implement logger.Interface
type zapLoggerWrapper struct {
	logger *zap.Logger
}

// NewZapLoggerWrapper creates a new wrapper around a zap logger
func NewZapLoggerWrapper(logger *zap.Logger) logger.Interface {
	return &zapLoggerWrapper{logger: logger}
}

// Debug implements logger.Interface
func (w *zapLoggerWrapper) Debug(msg string, fields ...any) {
	w.logger.Debug(msg, toZapFields(fields)...)
}

// Info implements logger.Interface
func (w *zapLoggerWrapper) Info(msg string, fields ...any) {
	w.logger.Info(msg, toZapFields(fields)...)
}

// Warn implements logger.Interface
func (w *zapLoggerWrapper) Warn(msg string, fields ...any) {
	w.logger.Warn(msg, toZapFields(fields)...)
}

// Error implements logger.Interface
func (w *zapLoggerWrapper) Error(msg string, fields ...any) {
	w.logger.Error(msg, toZapFields(fields)...)
}

// Fatal implements logger.Interface
func (w *zapLoggerWrapper) Fatal(msg string, fields ...any) {
	w.logger.Fatal(msg, toZapFields(fields)...)
}

// With implements logger.Interface
func (w *zapLoggerWrapper) With(fields ...any) logger.Interface {
	return &zapLoggerWrapper{logger: w.logger.With(toZapFields(fields)...)}
}

// toZapFields converts a list of any fields to zap.Field
func toZapFields(fields []any) []zap.Field {
	if len(fields)%2 != 0 {
		return []zap.Field{zap.Error(errors.New("invalid fields: must be key-value pairs"))}
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(errors.New("invalid fields: keys must be strings"))}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
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

// Start begins the crawling process for a given source.
func (c *Crawler) Start(ctx context.Context, sourceName string) error {
	// Validate source exists
	source := c.sources.FindByName(sourceName)
	if source == nil {
		return fmt.Errorf("source not found: %s", sourceName)
	}

	// Validate index exists
	exists, err := c.indexManager.IndexExists(ctx, sourceName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("index not found: %s", sourceName)
	}

	// Configure collector for this source
	c.configureCollector(source)

	// Start crawler in a goroutine
	go func() {
		if err := c.crawl(ctx, source); err != nil {
			c.logger.Error("Crawler error", zap.Error(err))
		}
	}()

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
	c.collector.OnHTML("article", func(e *colly.HTMLElement) {
		article := &models.Article{
			ID:        e.Request.URL.String(),
			Title:     e.ChildText("h1"),
			Body:      e.ChildText(".content"),
			Source:    c.state.currentSource,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		select {
		case c.articleChannel <- article:
			c.metrics.UpdateMetrics(true)
		default:
			c.logger.Warn("Article channel full, dropping article")
		}
	})

	c.collector.OnError(func(r *colly.Response, err error) {
		c.logger.Error("Crawler error",
			zap.String("url", r.Request.URL.String()),
			zap.Error(err),
		)
		c.metrics.UpdateMetrics(false)
	})
}

// configureCollector configures the collector with the given source configuration.
func (c *Crawler) configureCollector(source *sourceutils.SourceConfig) {
	c.collector = colly.NewCollector(
		colly.AllowedDomains(source.AllowedDomains...),
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
	)

	// Set rate limits
	c.collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: source.RateLimit,
		Parallelism: 1, // Default parallelism
	})

	// Set up callbacks
	c.setupCallbacks()
}

// crawl performs the actual crawling process.
func (c *Crawler) crawl(ctx context.Context, source *sourceutils.SourceConfig) error {
	c.state.mu.Lock()
	c.state.currentSource = source.Name
	c.state.isRunning = true
	c.state.mu.Unlock()

	defer func() {
		c.state.mu.Lock()
		c.state.isRunning = false
		c.state.mu.Unlock()
	}()

	// Start the crawler
	if err := c.collector.Visit(source.URL); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Wait for the crawler to finish
	c.collector.Wait()

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
