// Package page provides functionality for processing and managing web pages.
package page

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// PageProcessor implements the content.Processor interface for pages.
type PageProcessor struct {
	logger      logger.Interface
	service     Interface
	validator   content.JobValidator
	storage     types.Interface
	indexName   string
	pageChannel chan *domain.Page
	registry    []content.ContentProcessor
}

// NewPageProcessor creates a new page processor.
func NewPageProcessor(
	log logger.Interface,
	service Interface,
	validator content.JobValidator,
	storage types.Interface,
	indexName string,
	pageChannel chan *domain.Page,
) *PageProcessor {
	return &PageProcessor{
		logger:      log,
		service:     service,
		validator:   validator,
		storage:     storage,
		indexName:   indexName,
		pageChannel: pageChannel,
		registry:    make([]content.ContentProcessor, 0),
	}
}

// Process implements the content.Processor interface.
func (p *PageProcessor) Process(ctx context.Context, contentData any) error {
	e, ok := contentData.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", contentData)
	}

	// Process the page
	if err := p.service.Process(e); err != nil {
		return fmt.Errorf("failed to process page: %w", err)
	}

	// Send the processed page to the channel
	if p.pageChannel != nil {
		page := &domain.Page{
			URL:       e.Request.URL.String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		p.pageChannel <- page
	}

	return nil
}

// ContentType implements the content.Processor interface.
func (p *PageProcessor) ContentType() contenttype.Type {
	return contenttype.Page
}

// CanProcess implements the content.Processor interface.
func (p *PageProcessor) CanProcess(ct contenttype.Type) bool {
	return ct == contenttype.Page
}

// Start implements the content.Processor interface.
func (p *PageProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements the content.Processor interface.
func (p *PageProcessor) Stop(ctx context.Context) error {
	if p.pageChannel != nil {
		close(p.pageChannel)
	}
	return nil
}

// ValidateJob implements the content.Processor interface.
func (p *PageProcessor) ValidateJob(job *content.Job) error {
	if p.validator == nil {
		return nil
	}
	return p.validator.ValidateJob(job)
}

// GetProcessor returns a processor for the given content type.
func (p *PageProcessor) GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error) {
	if contentType == contenttype.Page {
		return &pageContentProcessor{p}, nil
	}

	for _, processor := range p.registry {
		if processor.CanProcess(contentType) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

// ParseHTML parses HTML content.
func (p *PageProcessor) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// ExtractLinks extracts links from the content.
func (p *PageProcessor) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// ExtractContent extracts the main content.
func (p *PageProcessor) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
}

// RegisterProcessor registers a new processor.
func (p *PageProcessor) RegisterProcessor(processor content.ContentProcessor) {
	p.registry = append(p.registry, processor)
}

// ProcessContent implements content.ProcessorRegistry
func (p *PageProcessor) ProcessContent(ctx context.Context, ct contenttype.Type, contentData any) error {
	proc, err := p.GetProcessor(ct)
	if err != nil {
		return err
	}
	return proc.Process(ctx, contentData)
}

// pageContentProcessor wraps PageProcessor to implement content.ContentProcessor
type pageContentProcessor struct {
	*PageProcessor
}

// Process implements content.ContentProcessor
func (p *pageContentProcessor) Process(ctx context.Context, contentData any) error {
	return p.PageProcessor.Process(ctx, contentData)
}

// ContentType implements content.ContentProcessor
func (p *pageContentProcessor) ContentType() contenttype.Type {
	return p.PageProcessor.ContentType()
}

// CanProcess implements content.ContentProcessor
func (p *pageContentProcessor) CanProcess(ct contenttype.Type) bool {
	return p.PageProcessor.CanProcess(ct)
}

// ValidateJob implements content.ContentProcessor
func (p *pageContentProcessor) ValidateJob(job *content.Job) error {
	return p.PageProcessor.ValidateJob(job)
}
