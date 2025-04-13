// Package common provides common types and interfaces used across the application.
package common

import (
	"context"
	"fmt"
	"io"

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
	// ContentTypeHTML represents HTML content
	ContentTypeHTML ContentType = "html"
	// ContentTypeJob represents job content
	ContentTypeJob ContentType = "job"
)

// ContentProcessor defines the interface for processing different types of content.
type ContentProcessor interface {
	// ContentType returns the type of content this processor can handle.
	ContentType() ContentType

	// CanProcess checks if the processor can handle the given content.
	CanProcess(content any) bool

	// Process handles the content processing.
	Process(ctx context.Context, content any) error
}

// HTMLProcessor defines the interface for processing HTML content.
type HTMLProcessor interface {
	ContentProcessor

	// ParseHTML parses HTML content from a reader.
	ParseHTML(r io.Reader) error

	// ExtractLinks extracts links from the parsed HTML.
	ExtractLinks() ([]string, error)

	// ExtractContent extracts the main content from the parsed HTML.
	ExtractContent() (string, error)
}

// JobProcessor defines the interface for processing jobs.
type JobProcessor interface {
	ContentProcessor

	// ProcessJob processes a job and its items.
	ProcessJob(ctx context.Context, job *Job) error

	// ValidateJob validates a job before processing.
	ValidateJob(job *Job) error
}

// ProcessorRegistry manages content processors
type ProcessorRegistry interface {
	// RegisterProcessor registers a new content processor.
	RegisterProcessor(processor ContentProcessor)

	// GetProcessor returns a processor for the given content type.
	GetProcessor(contentType ContentType) (ContentProcessor, error)

	// ProcessContent processes content using the appropriate processor.
	ProcessContent(ctx context.Context, contentType ContentType, content any) error
}

// Processor combines multiple processing capabilities.
type Processor interface {
	HTMLProcessor
	JobProcessor
	ProcessorRegistry
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

// CanProcess implements ContentProcessor.CanProcess
func (p *NoopProcessor) CanProcess(content any) bool {
	return true
}

// Process implements ContentProcessor.Process
func (p *NoopProcessor) Process(ctx context.Context, content any) error {
	return nil
}

// ParseHTML implements HTMLProcessor.ParseHTML
func (p *NoopProcessor) ParseHTML(r io.Reader) error {
	return nil
}

// ExtractLinks implements HTMLProcessor.ExtractLinks
func (p *NoopProcessor) ExtractLinks() ([]string, error) {
	return nil, nil
}

// ExtractContent implements HTMLProcessor.ExtractContent
func (p *NoopProcessor) ExtractContent() (string, error) {
	return "", nil
}

// ProcessJob implements JobProcessor.ProcessJob
func (p *NoopProcessor) ProcessJob(ctx context.Context, job *Job) error {
	return nil
}

// ValidateJob implements JobProcessor.ValidateJob
func (p *NoopProcessor) ValidateJob(job *Job) error {
	return nil
}

// RegisterProcessor implements ProcessorRegistry.RegisterProcessor
func (p *NoopProcessor) RegisterProcessor(processor ContentProcessor) {
	// No-op implementation
}

// GetProcessor implements ProcessorRegistry.GetProcessor
func (p *NoopProcessor) GetProcessor(contentType ContentType) (ContentProcessor, error) {
	return p, nil
}

// ProcessContent implements ProcessorRegistry.ProcessContent
func (p *NoopProcessor) ProcessContent(ctx context.Context, contentType ContentType, content any) error {
	return nil
}

// Start implements Processor.Start
func (p *NoopProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements Processor.Stop
func (p *NoopProcessor) Stop(ctx context.Context) error {
	return nil
}
