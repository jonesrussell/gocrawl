// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Handler defines the interface for content handlers.
type Handler interface {
	Process(ctx context.Context, element *colly.HTMLElement) error
}

// Validator defines the interface for content validation.
type Validator interface {
	Validate(element *colly.HTMLElement) error
}

// Processor implements the ContentProcessor interface.
type Processor struct {
	logger    logger.Interface
	metrics   CrawlerMetrics
	handlers  map[string]Handler
	validator Validator
}

// NewProcessor creates a new content processor.
func NewProcessor(
	logger logger.Interface,
	metrics CrawlerMetrics,
	validator Validator,
) ContentProcessor {
	return &Processor{
		logger:    logger,
		metrics:   metrics,
		handlers:  make(map[string]Handler),
		validator: validator,
	}
}

// ProcessHTML processes HTML content.
func (p *Processor) ProcessHTML(ctx context.Context, element *colly.HTMLElement) error {
	// Get content type from element
	contentType := p.detectContentType(element)
	if contentType == "" {
		return fmt.Errorf("unable to detect content type")
	}

	// Get handler for content type
	handler, ok := p.handlers[contentType]
	if !ok {
		return fmt.Errorf("no handler for content type: %s", contentType)
	}

	// Validate content
	if err := p.validator.Validate(element); err != nil {
		p.logger.Error("Content validation failed",
			"error", err,
			"url", element.Request.URL.String(),
			"content_type", contentType)
		p.metrics.IncrementError()
		return fmt.Errorf("content validation failed: %w", err)
	}

	// Process content
	if err := handler.Process(ctx, element); err != nil {
		p.logger.Error("Content processing failed",
			"error", err,
			"url", element.Request.URL.String(),
			"content_type", contentType)
		p.metrics.IncrementError()
		return fmt.Errorf("content processing failed: %w", err)
	}

	p.metrics.IncrementProcessed()
	return nil
}

// CanProcess returns whether the processor can handle the content.
func (p *Processor) CanProcess(contentType string) bool {
	_, ok := p.handlers[contentType]
	return ok
}

// ContentType returns the content type this processor handles.
func (p *Processor) ContentType() string {
	types := make([]string, 0, len(p.handlers))
	for t := range p.handlers {
		types = append(types, t)
	}
	return strings.Join(types, ", ")
}

// RegisterHandler registers a handler for a content type.
func (p *Processor) RegisterHandler(contentType string, handler Handler) {
	p.handlers[contentType] = handler
}

// detectContentType detects the content type of an HTML element.
func (p *Processor) detectContentType(element *colly.HTMLElement) string {
	// Check for article content
	if p.isArticle(element) {
		return "article"
	}

	// Check for content types based on element attributes
	if element.Attr("data-type") != "" {
		return element.Attr("data-type")
	}

	// Check for content types based on element classes
	classes := element.Attr("class")
	if classes != "" {
		for _, class := range strings.Split(classes, " ") {
			if strings.Contains(class, "article") {
				return "article"
			}
			if strings.Contains(class, "content") {
				return "content"
			}
		}
	}

	return ""
}

// isArticle checks if an element is an article.
func (p *Processor) isArticle(element *colly.HTMLElement) bool {
	// Check for article tag
	if element.Name == "article" {
		return true
	}

	// Check for article role
	if element.Attr("role") == "article" {
		return true
	}

	// Check for article class
	classes := element.Attr("class")
	if classes != "" {
		for _, class := range strings.Split(classes, " ") {
			if strings.Contains(class, "article") {
				return true
			}
		}
	}

	return false
}
