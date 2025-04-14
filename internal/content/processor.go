// Package content provides content processing types and interfaces.
package content

import (
	"context"
	"fmt"
	"io"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
)

// ContentProcessor defines the interface for processing different types of content.
type ContentProcessor interface {
	// ContentType returns the type of content this processor can handle.
	ContentType() contenttype.Type

	// CanProcess checks if the processor can handle the given content.
	CanProcess(content contenttype.Type) bool

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

// ProcessorRegistry manages content processors.
type ProcessorRegistry interface {
	// RegisterProcessor registers a new content processor.
	RegisterProcessor(processor ContentProcessor)

	// GetProcessor returns a processor for the given content type.
	GetProcessor(contentType contenttype.Type) (ContentProcessor, error)

	// ProcessContent processes content using the appropriate processor.
	ProcessContent(ctx context.Context, contentType contenttype.Type, content any) error
}

// Processor defines the interface for content processors.
type Processor interface {
	HTMLProcessor
	ProcessorRegistry

	// Start initializes the processor.
	Start(ctx context.Context) error

	// Stop cleans up the processor.
	Stop(ctx context.Context) error
}

// ContentTypeDetector detects the type of content.
type ContentTypeDetector interface {
	// Detect detects the content type of the given content.
	Detect(content any) (contenttype.Type, error)
}

// HTMLContentTypeDetector detects content types in HTML.
type HTMLContentTypeDetector struct {
	selectors map[contenttype.Type]string
}

// NewHTMLContentTypeDetector creates a new HTML content type detector.
func NewHTMLContentTypeDetector(selectors map[contenttype.Type]string) *HTMLContentTypeDetector {
	return &HTMLContentTypeDetector{
		selectors: selectors,
	}
}

// Detect implements ContentTypeDetector.Detect.
func (d *HTMLContentTypeDetector) Detect(content any) (contenttype.Type, error) {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return "", fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	for contentType, selector := range d.selectors {
		if e.DOM.Find(selector).Length() > 0 {
			return contentType, nil
		}
	}

	return contenttype.Page, nil
}

// ProcessingStep represents a step in a processing pipeline.
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

// ProcessorFactory creates processors based on configuration.
type ProcessorFactory interface {
	CreateProcessor(config ProcessorConfig) (ContentProcessor, error)
}

// NoopProcessor implements Processor with no-op implementations.
type NoopProcessor struct{}

// ContentType implements ContentProcessor.ContentType.
func (p *NoopProcessor) ContentType() contenttype.Type {
	return contenttype.Page
}

// CanProcess implements ContentProcessor.CanProcess.
func (p *NoopProcessor) CanProcess(content contenttype.Type) bool {
	return true
}

// Process implements ContentProcessor.Process.
func (p *NoopProcessor) Process(ctx context.Context, content any) error {
	return nil
}

// ParseHTML implements HTMLProcessor.ParseHTML.
func (p *NoopProcessor) ParseHTML(r io.Reader) error {
	return nil
}

// ExtractLinks implements HTMLProcessor.ExtractLinks.
func (p *NoopProcessor) ExtractLinks() ([]string, error) {
	return nil, nil
}

// ExtractContent implements HTMLProcessor.ExtractContent.
func (p *NoopProcessor) ExtractContent() (string, error) {
	return "", nil
}

// RegisterProcessor implements ProcessorRegistry.RegisterProcessor.
func (p *NoopProcessor) RegisterProcessor(processor ContentProcessor) {
	// No-op
}

// GetProcessor implements ProcessorRegistry.GetProcessor.
func (p *NoopProcessor) GetProcessor(contentType contenttype.Type) (ContentProcessor, error) {
	return p, nil
}

// ProcessContent implements ProcessorRegistry.ProcessContent.
func (p *NoopProcessor) ProcessContent(ctx context.Context, contentType contenttype.Type, content any) error {
	return nil
}

// Start implements Processor.Start.
func (p *NoopProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements Processor.Stop.
func (p *NoopProcessor) Stop(ctx context.Context) error {
	return nil
}
