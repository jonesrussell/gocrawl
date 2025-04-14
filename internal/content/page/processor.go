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
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// PageProcessor implements the common.Processor interface for pages.
type PageProcessor struct {
	logger      logger.Interface
	service     Interface
	validator   common.JobValidator
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
	// Check if content is a job
	job, ok := content.(*jobtypes.Job)
	if !ok {
		return fmt.Errorf("invalid content type: expected *jobtypes.Job, got %T", content)
	}

	// Create a new collector for this job
	c := colly.NewCollector()

	// Configure the collector
	c.OnHTML("body", func(e *colly.HTMLElement) {
		p.service.Process(e)
	})

	// Visit the URL
	if err := c.Visit(job.URL); err != nil {
		return fmt.Errorf("failed to visit URL: %w", err)
	}

	// Send the processed page to the channel
	if p.pageChannel != nil {
		page := &models.Page{
			URL:       job.URL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		p.pageChannel <- page
	}

	return nil
}

// Close implements the common.Processor interface.
func (p *PageProcessor) Close() error {
	if p.pageChannel != nil {
		close(p.pageChannel)
	}
	return nil
}

// CanProcess implements the common.Processor interface.
func (p *PageProcessor) CanProcess(contentType any) bool {
	ct, ok := contentType.(common.ContentType)
	return ok && ct == common.ContentTypePage
}

// ContentType implements the common.Processor interface.
func (p *PageProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// ExtractContent implements the common.Processor interface.
func (p *PageProcessor) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
}

// ExtractLinks implements the common.Processor interface.
func (p *PageProcessor) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// GetProcessor implements the common.Processor interface.
func (p *PageProcessor) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	if contentType != common.ContentTypePage {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return p, nil
}

// ParseHTML implements the common.Processor interface.
func (p *PageProcessor) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// ProcessContent implements the common.Processor interface.
func (p *PageProcessor) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	if contentType != common.ContentTypePage {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	return p.Process(ctx, content)
}

// ProcessJob implements the common.Processor interface.
func (p *PageProcessor) ProcessJob(ctx context.Context, job *jobtypes.Job) error {
	return p.Process(ctx, job)
}

// RegisterProcessor implements the common.Processor interface.
func (p *PageProcessor) RegisterProcessor(processor common.ContentProcessor) {
	// No-op for now
}

// Start implements the common.Processor interface.
func (p *PageProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements the common.Processor interface.
func (p *PageProcessor) Stop(ctx context.Context) error {
	return nil
}

// ValidateJob implements the common.Processor interface.
func (p *PageProcessor) ValidateJob(job *jobtypes.Job) error {
	return p.validator.ValidateJob(job)
}
