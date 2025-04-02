// Package content provides functionality for processing and managing content.
package content

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentProcessor handles content processing.
type ContentProcessor struct {
	// Logger for content processing operations
	Logger common.Logger
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

// ProcessJob processes a job and its items.
func (p *ContentProcessor) ProcessJob(ctx context.Context, job *common.Job) {
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
		return
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

	// Extract content data
	content := &models.Content{
		ID:        uuid.New().String(),
		Title:     e.ChildText("h1"),
		Body:      e.ChildText("article"),
		Type:      "page",
		URL:       e.Request.URL.String(),
		CreatedAt: time.Now(),
	}

	// Process the content using the ContentService
	processedContent := p.ContentService.Process(ctx, content.ID)
	if processedContent == "" {
		p.Logger.Error("Failed to process content",
			"component", "content/processor",
			"contentID", content.ID)
		p.metrics.ErrorCount++
		return errors.New("failed to process content: empty result")
	}

	// Store the processed content
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

// ProcessContent implements the collector.Processor interface
func (p *ContentProcessor) ProcessContent(e *colly.HTMLElement) {
	p.Logger.Debug("Processing content from HTML",
		"component", "content/processor",
		"url", e.Request.URL.String())

	content := p.ContentService.ExtractContent(e)
	if content == nil {
		p.Logger.Debug("No content extracted",
			"component", "content/processor",
			"url", e.Request.URL.String())
		return
	}

	p.Logger.Debug("Content extracted",
		"component", "content/processor",
		"url", e.Request.URL.String(),
		"title", content.Title)

	if err := p.Storage.IndexDocument(context.Background(), p.IndexName, content.ID, content); err != nil {
		p.Logger.Error("Failed to index content",
			"component", "content/processor",
			"contentID", content.ID,
			"error", err)
	}
}

// Process implements common.Processor
func (p *ContentProcessor) Process(ctx context.Context, data any) error {
	content, ok := data.(*models.Content)
	if !ok {
		return fmt.Errorf("invalid data type: expected *models.Content, got %T", data)
	}

	// Process the content using the ContentService
	if err := p.Storage.IndexDocument(ctx, p.IndexName, content.ID, content); err != nil {
		p.Logger.Error("Failed to index content",
			"component", "content/processor",
			"contentID", content.ID,
			"error", err)
		p.metrics.ErrorCount++
		return err
	}

	p.metrics.ProcessedCount++
	p.metrics.LastProcessedTime = time.Now()
	return nil
}

// CanProcess implements ContentProcessor.CanProcess
func (p *ContentProcessor) CanProcess(content interface{}) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// ContentType implements ContentProcessor.ContentType
func (p *ContentProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// Ensure ContentProcessor implements common.Processor
var _ common.Processor = (*ContentProcessor)(nil)
