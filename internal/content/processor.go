package content

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentProcessor handles the processing of non-article content
type ContentProcessor struct {
	service   Interface
	storage   storagetypes.Interface
	logger    common.Logger
	indexName string
	metrics   *common.Metrics
}

// NewProcessor creates a new content processor instance.
func NewProcessor(
	service Interface,
	storage storagetypes.Interface,
	logger common.Logger,
	indexName string,
) *ContentProcessor {
	return &ContentProcessor{
		service:   service,
		storage:   storage,
		logger:    logger,
		indexName: indexName,
		metrics:   &common.Metrics{},
	}
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
		p.logger.Warn("Job processing cancelled",
			"job_id", job.ID,
			"error", ctx.Err(),
		)
		p.metrics.ErrorCount++
		return
	default:
		// Process the job
		p.logger.Info("Processing job",
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

// ProcessHTML processes HTML content from a source.
func (p *ContentProcessor) ProcessHTML(e *colly.HTMLElement) error {
	start := time.Now()
	defer func() {
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	p.logger.Debug("Processing content",
		"component", "content/processor",
		"url", e.Request.URL.String(),
		"index", p.indexName)

	content := p.service.ExtractContent(e)
	if content == nil {
		p.logger.Debug("No content extracted",
			"component", "content/processor",
			"url", e.Request.URL.String())
		return nil
	}

	p.logger.Debug("Content extracted",
		"component", "content/processor",
		"url", e.Request.URL.String(),
		"title", content.Title)

	if err := p.storage.IndexDocument(context.Background(), p.indexName, content.ID, content); err != nil {
		p.logger.Error("Failed to index content",
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

// GetMetrics returns the current processing metrics.
func (p *ContentProcessor) GetMetrics() *common.Metrics {
	return p.metrics
}

// Ensure ContentProcessor implements common.Processor
var _ common.Processor = (*ContentProcessor)(nil)
