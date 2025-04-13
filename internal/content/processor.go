// Package content provides functionality for processing and managing content.
package content

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// PageProcessor handles page processing.
type PageProcessor struct {
	// Logger for page processing operations
	Logger logger.Interface
	// PageService for page operations
	PageService Interface
	// Storage for page persistence
	Storage storagetypes.Interface
	// IndexName is the name of the page index
	IndexName string
	// metrics holds processing metrics
	metrics *common.Metrics
}

// NewPageProcessor creates a new page processor.
func NewPageProcessor(p ProcessorParams) *PageProcessor {
	return &PageProcessor{
		Logger:      p.Logger,
		PageService: p.Service,
		Storage:     p.Storage,
		IndexName:   p.IndexName,
		metrics:     &common.Metrics{},
	}
}

// Start implements common.Processor.Start.
func (p *PageProcessor) Start(ctx context.Context) error {
	p.Logger.Info("Starting page processor",
		"component", "page/processor")
	return nil
}

// Stop implements common.Processor.Stop.
func (p *PageProcessor) Stop(ctx context.Context) error {
	p.Logger.Info("Stopping page processor",
		"component", "page/processor")
	return nil
}

// ProcessJob implements common.Processor.ProcessJob
func (p *PageProcessor) ProcessJob(ctx context.Context, job *common.Job) error {
	start := time.Now()
	defer func() {
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		p.Logger.Warn("Job processing cancelled",
			"job_id", job.ID,
			"error", ctx.Err(),
		)
		p.metrics.ErrorCount++
		return ctx.Err()
	default:
		// Process the job
		p.Logger.Info("Processing job",
			"job_id", job.ID,
		)

		// TODO: Implement job processing logic
		// This would typically involve:
		// 1. Fetching items associated with the job
		// 2. Processing each item
		// 3. Updating job status
		// 4. Handling errors and retries

		p.metrics.ProcessedCount++
		return nil
	}
}

// ProcessHTML implements HTMLProcessor.ProcessHTML
func (p *PageProcessor) ProcessHTML(ctx context.Context, e *colly.HTMLElement) error {
	start := time.Now()
	defer func() {
		p.metrics.LastProcessedTime = time.Now()
		p.metrics.ProcessedCount++
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	// Extract page using the PageService
	page := p.PageService.ExtractContent(e)
	if page == nil {
		p.Logger.Debug("No page extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	page.Metadata = p.PageService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, typeExists := page.Metadata["@type"].(string); typeExists {
		jsonLDType = typeVal
	}

	// Determine page type
	page.Type = string(p.PageService.DetermineContentType(
		e.Request.URL.String(),
		page.Metadata,
		jsonLDType,
	))

	// Process page
	page.Body = p.PageService.Process(ctx, page.Body)

	// Store the page
	if err := p.Storage.IndexDocument(ctx, p.IndexName, page.ID, page); err != nil {
		p.Logger.Error("Failed to index page",
			"component", "page/processor",
			"pageID", page.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	return nil
}

// GetMetrics returns the current processing metrics.
func (p *PageProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// ProcessContent implements common.Processor.ProcessContent
func (p *PageProcessor) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	// Handle both page and image content types
	if contentType != common.ContentTypePage && contentType != common.ContentTypeImage {
		return fmt.Errorf("unsupported content type: %v", contentType)
	}

	e, isHTMLElement := content.(*colly.HTMLElement)
	if !isHTMLElement {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Extract page using the PageService
	page := p.PageService.ExtractContent(e)
	if page == nil {
		p.Logger.Debug("No page extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	page.Metadata = p.PageService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, typeExists := page.Metadata["@type"].(string); typeExists {
		jsonLDType = typeVal
	}

	// Determine page type
	page.Type = string(p.PageService.DetermineContentType(
		e.Request.URL.String(),
		page.Metadata,
		jsonLDType,
	))

	// Process page
	page.Body = p.PageService.Process(ctx, page.Body)

	// Store the page
	if err := p.Storage.IndexDocument(ctx, p.IndexName, page.ID, page); err != nil {
		p.Logger.Error("Failed to index page",
			"component", "page/processor",
			"pageID", page.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	return nil
}

// Process processes the page and returns the processed result
func (p *PageProcessor) Process(ctx context.Context, content any) error {
	e, isHTMLElement := content.(*colly.HTMLElement)
	if !isHTMLElement {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Extract page using the PageService
	page := p.PageService.ExtractContent(e)
	if page == nil {
		p.Logger.Debug("No page extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	page.Metadata = p.PageService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, typeExists := page.Metadata["@type"].(string); typeExists {
		jsonLDType = typeVal
	}

	// Determine page type
	page.Type = string(p.PageService.DetermineContentType(
		e.Request.URL.String(),
		page.Metadata,
		jsonLDType,
	))

	// Process page
	page.Body = p.PageService.Process(ctx, page.Body)

	// Store the page
	if err := p.Storage.IndexDocument(ctx, p.IndexName, page.ID, page); err != nil {
		p.Logger.Error("Failed to index page",
			"component", "page/processor",
			"pageID", page.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	p.metrics.ProcessedCount++
	p.metrics.LastProcessedTime = time.Now()
	return nil
}

// CanProcess checks if this processor can handle the given content
func (p *PageProcessor) CanProcess(content any) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// ContentType implements ContentProcessor.ContentType
func (p *PageProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// ExtractContent implements common.Processor.ExtractContent
func (p *PageProcessor) ExtractContent() (string, error) {
	// For page processing, we extract the raw content from the HTML element
	// This is a no-op implementation since we process content directly in ProcessHTML
	return "", nil
}

// ExtractLinks implements common.Processor.ExtractLinks
func (p *PageProcessor) ExtractLinks() ([]string, error) {
	// For page processing, we don't need to extract links
	// as we process the content directly
	return nil, nil
}

// GetProcessor implements common.Processor.GetProcessor
func (p *PageProcessor) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	// For page processing, we handle page content
	if contentType != common.ContentTypePage {
		return nil, fmt.Errorf("unsupported content type: %v", contentType)
	}
	return p, nil
}

// ParseHTML implements common.Processor.ParseHTML
func (p *PageProcessor) ParseHTML(r io.Reader) error {
	// For page processing, we don't need to parse raw HTML
	// as we process the content directly in ProcessHTML
	return nil
}

// RegisterProcessor implements common.Processor.RegisterProcessor
func (p *PageProcessor) RegisterProcessor(processor common.ContentProcessor) {
	// For page processing, we don't need to register additional processors
	// as we handle all content types directly
}

// ValidateJob implements common.Processor.ValidateJob
func (p *PageProcessor) ValidateJob(job *common.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if job.ID == "" {
		return errors.New("job ID cannot be empty")
	}
	return nil
}

// Ensure PageProcessor implements common.Processor
var _ common.Processor = (*PageProcessor)(nil)
