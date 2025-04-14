// Package page provides functionality for processing and managing web pages.
package page

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/contenttype"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// PageProcessor implements the common.Processor interface for pages.
type PageProcessor struct {
	logger      logger.Interface
	service     Interface
	validator   jobtypes.JobValidator
	storage     types.Interface
	indexName   string
	pageChannel chan *models.Page
}

// NewPageProcessor creates a new page processor.
func NewPageProcessor(p ProcessorParams) *PageProcessor {
	return &PageProcessor{
		logger:      p.Logger,
		service:     p.Service,
		validator:   p.Validator,
		storage:     p.Storage,
		indexName:   p.IndexName,
		pageChannel: p.PageChannel,
	}
}

// Process implements the common.Processor interface.
func (p *PageProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Process the page
	if err := p.service.Process(e); err != nil {
		return fmt.Errorf("failed to process page: %w", err)
	}

	// Send the processed page to the channel
	if p.pageChannel != nil {
		page := &models.Page{
			URL:       e.Request.URL.String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		p.pageChannel <- page
	}

	return nil
}

// ContentType implements the common.Processor interface.
func (p *PageProcessor) ContentType() contenttype.Type {
	return contenttype.Page
}

// CanProcess implements the common.Processor interface.
func (p *PageProcessor) CanProcess(contentType contenttype.Type) bool {
	return contentType == contenttype.Page
}

// Start implements the common.Processor interface.
func (p *PageProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements the common.Processor interface.
func (p *PageProcessor) Stop(ctx context.Context) error {
	if p.pageChannel != nil {
		close(p.pageChannel)
	}
	return nil
}

// ValidateJob implements the common.Processor interface.
func (p *PageProcessor) ValidateJob(job *jobtypes.Job) error {
	if p.validator == nil {
		return nil
	}
	return p.validator.ValidateJob(job)
}

// GetProcessor returns a processor for the given content type.
func (p *PageProcessor) GetProcessor(contentType contenttype.Type) (common.Processor, error) {
	if contentType != contenttype.Page {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return p, nil
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
func (p *PageProcessor) RegisterProcessor(processor common.Processor) {
	// No-op for now
}
