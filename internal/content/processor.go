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

// ContentProcessor handles content processing.
type ContentProcessor struct {
	// Logger for content processing operations
	Logger logger.Interface
	// ContentService for content operations
	ContentService Interface
	// Storage for content persistence
	Storage storagetypes.Interface
	// IndexName is the name of the content index
	IndexName string
	// metrics holds processing metrics
	metrics *common.Metrics
}

// NewContentProcessor creates a new content processor.
func NewContentProcessor(p ProcessorParams) *ContentProcessor {
	return &ContentProcessor{
		Logger:         p.Logger,
		ContentService: p.Service,
		Storage:        p.Storage,
		IndexName:      p.IndexName,
		metrics:        &common.Metrics{},
	}
}

// Start implements common.Processor.Start.
func (p *ContentProcessor) Start(ctx context.Context) error {
	p.Logger.Info("Starting content processor",
		"component", "content/processor")
	return nil
}

// Stop implements common.Processor.Stop.
func (p *ContentProcessor) Stop(ctx context.Context) error {
	p.Logger.Info("Stopping content processor",
		"component", "content/processor")
	return nil
}

// ProcessJob implements common.Processor.ProcessJob
func (p *ContentProcessor) ProcessJob(ctx context.Context, job *common.Job) error {
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
func (p *ContentProcessor) ProcessHTML(ctx context.Context, e *colly.HTMLElement) error {
	start := time.Now()
	defer func() {
		p.metrics.LastProcessedTime = time.Now()
		p.metrics.ProcessedCount++
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	// Extract content using the ContentService
	content := p.ContentService.ExtractContent(e)
	if content == nil {
		p.Logger.Debug("No content extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	content.Metadata = p.ContentService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, ok := content.Metadata["@type"].(string); ok {
		jsonLDType = typeVal
	}

	// Determine content type
	content.Type = string(p.ContentService.DetermineContentType(
		e.Request.URL.String(),
		content.Metadata,
		jsonLDType,
	))

	// Process content
	content.Body = p.ContentService.Process(ctx, content.Body)

	// Store the content
	if err := p.Storage.IndexDocument(ctx, p.IndexName, content.ID, content); err != nil {
		p.Logger.Error("Failed to index content",
			"component", "content/processor",
			"contentID", content.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	return nil
}

// GetMetrics returns the current processing metrics.
func (p *ContentProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// ProcessContent implements common.Processor.ProcessContent
func (p *ContentProcessor) ProcessContent(ctx context.Context, contentType common.ContentType, content any) error {
	// Handle both page and image content types
	if contentType != common.ContentTypePage && contentType != common.ContentTypeImage {
		return fmt.Errorf("unsupported content type: %v", contentType)
	}

	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Extract content using the ContentService
	contentData := p.ContentService.ExtractContent(e)
	if contentData == nil {
		p.Logger.Debug("No content extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	contentData.Metadata = p.ContentService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, ok := contentData.Metadata["@type"].(string); ok {
		jsonLDType = typeVal
	}

	// Determine content type
	contentData.Type = string(p.ContentService.DetermineContentType(
		e.Request.URL.String(),
		contentData.Metadata,
		jsonLDType,
	))

	// Process content
	contentData.Body = p.ContentService.Process(ctx, contentData.Body)

	// Store the content
	if err := p.Storage.IndexDocument(ctx, p.IndexName, contentData.ID, contentData); err != nil {
		p.Logger.Error("Failed to index content",
			"component", "content/processor",
			"contentID", contentData.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	return nil
}

// Process processes the content and returns the processed result
func (p *ContentProcessor) Process(ctx context.Context, content any) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}

	// Extract content using the ContentService
	contentData := p.ContentService.ExtractContent(e)
	if contentData == nil {
		p.Logger.Debug("No content extracted",
			"url", e.Request.URL.String())
		return nil
	}

	// Extract metadata
	contentData.Metadata = p.ContentService.ExtractMetadata(e)

	// Get JSON-LD type if available
	jsonLDType := ""
	if typeVal, ok := contentData.Metadata["@type"].(string); ok {
		jsonLDType = typeVal
	}

	// Determine content type
	contentData.Type = string(p.ContentService.DetermineContentType(
		e.Request.URL.String(),
		contentData.Metadata,
		jsonLDType,
	))

	// Process content
	contentData.Body = p.ContentService.Process(ctx, contentData.Body)

	// Store the content
	if err := p.Storage.IndexDocument(ctx, p.IndexName, contentData.ID, contentData); err != nil {
		p.Logger.Error("Failed to index content",
			"component", "content/processor",
			"contentID", contentData.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	p.metrics.ProcessedCount++
	p.metrics.LastProcessedTime = time.Now()
	return nil
}

// CanProcess checks if this processor can handle the given content
func (p *ContentProcessor) CanProcess(content any) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// ContentType implements ContentProcessor.ContentType
func (p *ContentProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// ExtractContent implements common.Processor.ExtractContent
func (p *ContentProcessor) ExtractContent() (string, error) {
	// For content processing, we extract the raw content from the HTML element
	// This is a no-op implementation since we process content directly in ProcessHTML
	return "", nil
}

// ExtractLinks implements common.Processor.ExtractLinks
func (p *ContentProcessor) ExtractLinks() ([]string, error) {
	// For content processing, we don't need to extract links
	// as we process the content directly
	return nil, nil
}

// GetProcessor implements common.Processor.GetProcessor
func (p *ContentProcessor) GetProcessor(contentType common.ContentType) (common.ContentProcessor, error) {
	// For content processing, we handle page content
	if contentType != common.ContentTypePage {
		return nil, fmt.Errorf("unsupported content type: %v", contentType)
	}
	return p, nil
}

// ParseHTML implements common.Processor.ParseHTML
func (p *ContentProcessor) ParseHTML(r io.Reader) error {
	// For content processing, we don't need to parse raw HTML
	// as we process the content directly in ProcessHTML
	return nil
}

// RegisterProcessor implements common.Processor.RegisterProcessor
func (p *ContentProcessor) RegisterProcessor(processor common.ContentProcessor) {
	// For content processing, we don't need to register additional processors
	// as we handle all content types directly
}

// ValidateJob implements common.Processor.ValidateJob
func (p *ContentProcessor) ValidateJob(job *common.Job) error {
	if job == nil {
		return errors.New("job cannot be nil")
	}
	if job.ID == "" {
		return errors.New("job ID cannot be empty")
	}
	return nil
}

// Ensure ContentProcessor implements common.Processor
var _ common.Processor = (*ContentProcessor)(nil)
