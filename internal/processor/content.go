package processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// ContentProcessor processes content from sources.
type ContentProcessor struct {
	Logger         common.Logger
	Storage        common.Storage
	IndexName      string
	ContentService content.Interface
	metrics        *common.Metrics
}

// NewContentProcessor creates a new content processor.
func NewContentProcessor(
	logger common.Logger,
	storage common.Storage,
	indexName string,
	contentService content.Interface,
) *ContentProcessor {
	return &ContentProcessor{
		Logger:         logger,
		Storage:        storage,
		IndexName:      indexName,
		ContentService: contentService,
		metrics:        &common.Metrics{},
	}
}

// ContentType implements ContentProcessor.ContentType
func (p *ContentProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// CanProcess implements ContentProcessor.CanProcess
func (p *ContentProcessor) CanProcess(content interface{}) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// Process implements ContentProcessor.Process
func (p *ContentProcessor) Process(ctx context.Context, content interface{}) error {
	e, ok := content.(*colly.HTMLElement)
	if !ok {
		return fmt.Errorf("invalid content type: expected *colly.HTMLElement, got %T", content)
	}
	return p.ProcessHTML(ctx, e)
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

// GetMetrics returns the current processor metrics.
func (p *ContentProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// Ensure ContentProcessor implements required interfaces
var (
	_ common.ContentProcessor = (*ContentProcessor)(nil)
	_ common.HTMLProcessor    = (*ContentProcessor)(nil)
)

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

// Start implements Processor.Start
func (p *ContentProcessor) Start(ctx context.Context) error {
	p.Logger.Info("Starting content processor")
	return nil
}

// Stop implements Processor.Stop
func (p *ContentProcessor) Stop(ctx context.Context) error {
	p.Logger.Info("Stopping content processor")
	return nil
}
