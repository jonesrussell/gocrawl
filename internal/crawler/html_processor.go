// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// HTMLProcessor processes HTML content and delegates to appropriate content processors.
type HTMLProcessor struct {
	logger       logger.Interface
	processors   []content.Processor
	unknownTypes map[contenttype.Type]int
}

// NewHTMLProcessor creates a new HTMLProcessor.
func NewHTMLProcessor(logger logger.Interface) *HTMLProcessor {
	return &HTMLProcessor{
		logger:       logger,
		processors:   make([]content.Processor, 0, 2), // Pre-allocate for article and page processors
		unknownTypes: make(map[contenttype.Type]int),
	}
}

// Process processes an HTML element.
func (p *HTMLProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Detect content type
	contentType := p.detectContentType(e)

	// Select appropriate processor
	processor := p.selectProcessor(contentType)
	if processor == nil {
		p.unknownTypes[contentType]++
		return fmt.Errorf("no processor found for content type: %s", contentType)
	}

	// Process the content
	if err := processor.Process(ctx, e); err != nil {
		return fmt.Errorf("failed to process content: %w", err)
	}

	return nil
}

// ParseHTML parses HTML content.
func (p *HTMLProcessor) ParseHTML(r io.Reader) error {
	return fmt.Errorf("not implemented")
}

// ExtractLinks extracts links from the content.
func (p *HTMLProcessor) ExtractLinks() ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// ExtractContent extracts the main content.
func (p *HTMLProcessor) ExtractContent() (string, error) {
	return "", fmt.Errorf("not implemented")
}

// CanProcess returns whether the processor can handle the given content type.
func (p *HTMLProcessor) CanProcess(contentType contenttype.Type) bool {
	return contentType == contenttype.HTML
}

// ContentType returns the content type this processor handles.
func (p *HTMLProcessor) ContentType() contenttype.Type {
	return contenttype.HTML
}

// Start initializes the processor.
func (p *HTMLProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop stops the processor.
func (p *HTMLProcessor) Stop(ctx context.Context) error {
	return nil
}

// ValidateJob validates a job before processing.
func (p *HTMLProcessor) ValidateJob(job *content.Job) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	return nil
}

// GetProcessor returns a processor for the given content type.
func (p *HTMLProcessor) GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error) {
	for _, processor := range p.processors {
		if processor.CanProcess(contentType) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no processor found for content type: %s", contentType)
}

// RegisterProcessor registers a new processor.
func (p *HTMLProcessor) RegisterProcessor(processor content.Processor) {
	p.processors = append(p.processors, processor)
}

// ProcessContent processes content using the appropriate processor.
func (p *HTMLProcessor) ProcessContent(ctx context.Context, contentType contenttype.Type, content any) error {
	processor, err := p.GetProcessor(contentType)
	if err != nil {
		return err
	}
	return processor.Process(ctx, content)
}

// selectProcessor selects a processor for the given content type.
func (p *HTMLProcessor) selectProcessor(contentType contenttype.Type) content.Processor {
	for _, processor := range p.processors {
		if processor.CanProcess(contentType) {
			return processor
		}
	}
	return nil
}

// detectContentType detects the content type of the given HTML element.
func (p *HTMLProcessor) detectContentType(e *colly.HTMLElement) contenttype.Type {
	url := e.Request.URL.String()

	// Check for article patterns
	if strings.Contains(url, "/article/") || strings.Contains(url, "/articles/") ||
		strings.Contains(url, "/blog/") || strings.Contains(url, "/news/") {
		return contenttype.Article
	}

	// Check for job listings
	if strings.Contains(url, "/job/") || strings.Contains(url, "/jobs/") ||
		strings.Contains(url, "/career/") || strings.Contains(url, "/careers/") {
		return contenttype.Job
	}

	// Check for image galleries
	if strings.Contains(url, "/image/") || strings.Contains(url, "/images/") ||
		strings.Contains(url, "/photo/") || strings.Contains(url, "/photos/") ||
		strings.Contains(url, "/gallery/") {
		return contenttype.Image
	}

	// Default to page
	return contenttype.Page
}

// GetUnknownTypes returns a map of content types that have no registered processor.
func (p *HTMLProcessor) GetUnknownTypes() map[contenttype.Type]int {
	return p.unknownTypes
}
