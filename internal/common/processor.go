// Package common provides common types and interfaces used across the application.
package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// ContentType represents the type of content being processed
type ContentType string

const (
	// ContentTypeArticle represents article content
	ContentTypeArticle ContentType = "article"
	// ContentTypePage represents generic page content
	ContentTypePage ContentType = "page"
	// ContentTypeVideo represents video content
	ContentTypeVideo ContentType = "video"
	// ContentTypeImage represents image content
	ContentTypeImage ContentType = "image"
)

// ContentProcessor defines the base interface for all content processors
type ContentProcessor interface {
	// ContentType returns the type of content this processor handles
	ContentType() ContentType
	// CanProcess checks if this processor can handle the given content
	CanProcess(content any) bool
	// Process processes the content and returns the processed result
	Process(ctx context.Context, content any) error
	// GetMetrics returns processing metrics
	GetMetrics() *Metrics
}

// HTMLProcessor extends ContentProcessor for HTML content
type HTMLProcessor interface {
	ContentProcessor
	// ProcessHTML processes HTML content
	ProcessHTML(ctx context.Context, e *colly.HTMLElement) error
}

// JobProcessor extends ContentProcessor for job-based processing
type JobProcessor interface {
	ContentProcessor
	// ProcessJob processes a job and its items
	ProcessJob(ctx context.Context, job *Job)
}

// ProcessorRegistry manages content processors
type ProcessorRegistry struct {
	processors map[ContentType]ContentProcessor
	mu         sync.RWMutex
}

// NewProcessorRegistry creates a new processor registry
func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[ContentType]ContentProcessor),
	}
}

// RegisterProcessor registers a new content processor
func (r *ProcessorRegistry) RegisterProcessor(p ContentProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[p.ContentType()] = p
}

// GetProcessor returns a processor for the given content type
func (r *ProcessorRegistry) GetProcessor(contentType ContentType) (ContentProcessor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.processors[contentType]
	return p, ok
}

// Processor combines ContentProcessor, HTMLProcessor, and JobProcessor with lifecycle methods
type Processor interface {
	ContentProcessor
	HTMLProcessor
	JobProcessor
	// Start initializes the processor
	Start(ctx context.Context) error
	// Stop cleans up the processor
	Stop(ctx context.Context) error
}

// ContentTypeDetector defines the interface for content type detection.
type ContentTypeDetector interface {
	// Detect detects the content type of the given content.
	Detect(content any) (ContentType, error)
}

// HTMLContentTypeDetector implements ContentTypeDetector for HTML content.
type HTMLContentTypeDetector struct {
	selectors map[ContentType]string
}

// NewHTMLContentTypeDetector creates a new HTML content type detector
func NewHTMLContentTypeDetector(selectors map[ContentType]string) *HTMLContentTypeDetector {
	return &HTMLContentTypeDetector{
		selectors: selectors,
	}
}

// Detect implements ContentTypeDetector.
func (d *HTMLContentTypeDetector) Detect(content any) (ContentType, error) {
	e, ok := content.(colly.HTMLElement)
	if !ok {
		return "", fmt.Errorf("unsupported content type: %T", content)
	}

	// Check selectors for each content type
	for contentType, selector := range d.selectors {
		if e.DOM.Find(selector).Length() > 0 {
			return contentType, nil
		}
	}

	return ContentTypePage, nil // Default to page type
}

// ProcessingStep defines a step in a processing pipeline.
type ProcessingStep interface {
	// Process processes the content and returns the processed result.
	Process(ctx context.Context, content any) (any, error)
}

// ProcessingPipeline represents a pipeline of processing steps.
type ProcessingPipeline struct {
	steps []ProcessingStep
}

// Execute executes the pipeline on the given content.
func (p *ProcessingPipeline) Execute(ctx context.Context, content any) (any, error) {
	var err error
	for _, step := range p.steps {
		content, err = step.Process(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("step failed: %w", err)
		}
	}
	return content, nil
}

// ProcessorConfig holds configuration for a processor.
type ProcessorConfig struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Enabled bool           `json:"enabled"`
	Options map[string]any `json:"options"`
}

// ProcessorFactory creates processors based on configuration
type ProcessorFactory interface {
	CreateProcessor(config ProcessorConfig) (ContentProcessor, error)
}

// NoopProcessor implements Processor with no-op implementations.
type NoopProcessor struct{}

// ContentType implements ContentProcessor.ContentType
func (p *NoopProcessor) ContentType() ContentType {
	return ContentTypePage
}

// CanProcess implements Processor.CanProcess
func (p *NoopProcessor) CanProcess(content any) bool {
	return true
}

// Process implements Processor.Process
func (p *NoopProcessor) Process(ctx context.Context, content any) error {
	return nil
}

// ProcessHTML implements HTMLProcessor.ProcessHTML
func (p *NoopProcessor) ProcessHTML(ctx context.Context, e *colly.HTMLElement) error {
	return nil
}

// ProcessJob implements JobProcessor.ProcessJob
func (p *NoopProcessor) ProcessJob(ctx context.Context, job *Job) {
	// No-op implementation
}

// GetMetrics implements ContentProcessor.GetMetrics
func (p *NoopProcessor) GetMetrics() *Metrics {
	return &Metrics{
		ProcessedCount:     0,
		ErrorCount:         0,
		LastProcessedTime:  time.Time{},
		ProcessingDuration: 0,
	}
}

// Start implements Processor.Start
func (p *NoopProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements Processor.Stop
func (p *NoopProcessor) Stop(ctx context.Context) error {
	return nil
}
