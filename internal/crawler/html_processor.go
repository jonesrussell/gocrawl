// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/constants"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

// HTMLProcessor processes HTML content and delegates to appropriate content processors.
type HTMLProcessor struct {
	logger       logger.Interface
	processors   []content.Processor
	unknownTypes map[contenttype.Type]int
	sources      sources.Interface
}

// NewHTMLProcessor creates a new HTMLProcessor.
func NewHTMLProcessor(log logger.Interface, sourcesManager sources.Interface) *HTMLProcessor {
	return &HTMLProcessor{
		logger:       log,
		processors:   make([]content.Processor, 0, DefaultProcessorsCapacity), // Pre-allocate for article and page processors
		unknownTypes: make(map[contenttype.Type]int),
		sources:      sourcesManager,
	}
}

// Process processes an HTML element.
// Note: This method is part of the content.Processor interface but content type
// detection is handled by the crawler via detectContentType, not through this method.
func (p *HTMLProcessor) Process(ctx context.Context, contentData any) error {
	// This method is not used in the current implementation.
	// Content type detection happens in crawler.selectProcessor via detectContentType.
	return errors.New("not implemented")
}

// ParseHTML parses HTML content.
func (p *HTMLProcessor) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// ExtractLinks extracts links from the content.
func (p *HTMLProcessor) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// ExtractContent extracts the main content.
func (p *HTMLProcessor) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
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
		return errors.New("job cannot be nil")
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
func (p *HTMLProcessor) ProcessContent(ctx context.Context, ct contenttype.Type, contentData any) error {
	proc, err := p.GetProcessor(ct)
	if err != nil {
		return err
	}
	return proc.Process(ctx, contentData)
}

// DetectContentType detects the content type of the given HTML element using selector-based detection.
func (p *HTMLProcessor) DetectContentType(e *colly.HTMLElement, source *types.Source) contenttype.Type {
	// e.DOM is a goquery.Selection, and since OnHTML("html") is used,
	// e.DOM represents the html element, so Find() searches the entire document

	// Strategy 1: Check Open Graph type metadata
	ogType := e.DOM.Find("meta[property='og:type']").AttrOr("content", "")
	if ogType == "article" {
		p.logger.Debug("Detected article via og:type metadata")
		return contenttype.Article
	}

	// Strategy 2: Use article selectors to detect content
	// If the page matches article selectors and has substantial content, it's an article
	if source == nil || source.Selectors.Article.Body == "" {
		p.logger.Debug("No source or article body selector defined, defaulting to page")
		return contenttype.Page
	}

	// Get article body using the source's body selector
	bodySelector := source.Selectors.Article.Body
	articleBody := e.DOM.Find(bodySelector)
	if articleBody.Length() == 0 {
		p.logger.Debug("No article body found with selector", "selector", bodySelector)
		return contenttype.Page
	}

	// Verify it has substantial content (articles typically have >200 chars)
	bodyText := strings.TrimSpace(articleBody.Text())
	if len(bodyText) < constants.MinArticleBodyLength {
		p.logger.Debug("Body content too short, treating as page",
			"length", len(bodyText),
			"min_required", constants.MinArticleBodyLength)
		return contenttype.Page
	}

	// Strategy 3: Verify title exists (articles should have titles)
	titleSelector := source.Selectors.Article.Title
	if titleSelector != "" {
		articleTitle := e.DOM.Find(titleSelector)
		if articleTitle.Length() == 0 {
			p.logger.Debug("No article title found, treating as page")
			return contenttype.Page
		}

		titleText := strings.TrimSpace(articleTitle.Text())
		if titleText == "" {
			p.logger.Debug("Empty article title, treating as page")
			return contenttype.Page
		}
	}

	// If we got here, it has body + title with substantial content
	p.logger.Debug("Detected article via selectors", "body_length", len(bodyText))
	return contenttype.Article
}

// GetUnknownTypes returns a map of content types that have no registered processor.
func (p *HTMLProcessor) GetUnknownTypes() map[contenttype.Type]int {
	return p.unknownTypes
}
