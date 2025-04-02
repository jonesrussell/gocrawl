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
	CanProcess(content interface{}) bool
	// Process processes the content and returns the processed result
	Process(ctx context.Context, content interface{}) error
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

// ContentTypeDetector determines the type of content
type ContentTypeDetector interface {
	Detect(content interface{}) (ContentType, error)
}

// HTMLContentTypeDetector implements content type detection for HTML
type HTMLContentTypeDetector struct {
	selectors map[ContentType]string
}

// NewHTMLContentTypeDetector creates a new HTML content type detector
func NewHTMLContentTypeDetector(selectors map[ContentType]string) *HTMLContentTypeDetector {
	return &HTMLContentTypeDetector{
		selectors: selectors,
	}
}

// Detect determines the content type from HTML content
func (d *HTMLContentTypeDetector) Detect(content interface{}) (ContentType, error) {
	e, ok := content.(*colly.HTMLElement)
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

// ProcessingPipeline represents a chain of processing steps
type ProcessingPipeline struct {
	steps []ProcessingStep
}

// ProcessingStep represents a single step in the pipeline
type ProcessingStep interface {
	Process(ctx context.Context, content interface{}) (interface{}, error)
	Name() string
}

// AddPipelineStep adds a new step to the pipeline
func (p *ProcessingPipeline) AddStep(step ProcessingStep) {
	p.steps = append(p.steps, step)
}

// Execute runs the pipeline on the given content
func (p *ProcessingPipeline) Execute(ctx context.Context, content interface{}) (interface{}, error) {
	var err error
	for _, step := range p.steps {
		content, err = step.Process(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("step %s failed: %w", step.Name(), err)
		}
	}
	return content, nil
}

// ProcessorConfig holds configuration for content processors
type ProcessorConfig struct {
	ContentTypes []ContentType          `json:"content_types"`
	Selectors    map[string]string      `json:"selectors"`
	Options      map[string]interface{} `json:"options"`
}

// ProcessorFactory creates processors based on configuration
type ProcessorFactory interface {
	CreateProcessor(config ProcessorConfig) (ContentProcessor, error)
}

// NoopProcessor is a no-op implementation of ContentProcessor
type NoopProcessor struct{}

// NewNoopProcessor creates a new no-op processor
func NewNoopProcessor() *NoopProcessor {
	return &NoopProcessor{}
}

// ContentType implements ContentProcessor.ContentType
func (p *NoopProcessor) ContentType() ContentType {
	return ContentTypePage
}

// CanProcess implements ContentProcessor.CanProcess
func (p *NoopProcessor) CanProcess(content interface{}) bool {
	return true
}

// Process implements ContentProcessor.Process
func (p *NoopProcessor) Process(ctx context.Context, content interface{}) error {
	return nil
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
