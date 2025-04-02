package processor

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ContentService handles content extraction and processing.
type ContentService struct {
	logger common.Logger
}

// NewContentService creates a new content service.
func NewContentService(logger common.Logger) *ContentService {
	return &ContentService{
		logger: logger,
	}
}

// ContentProcessor implements the collector.Processor interface for general content.
type ContentProcessor struct {
	service   *ContentService
	storage   types.Interface
	logger    common.Logger
	indexName string

	// Metrics
	processedCount    int64
	errorCount        int64
	lastProcessedTime time.Time
	processingTime    time.Duration
}

// Process implements common.Processor
func (p *ContentProcessor) Process(e *colly.HTMLElement) error {
	return p.ProcessHTML(e)
}

// ProcessHTML implements common.Processor
func (p *ContentProcessor) ProcessHTML(e *colly.HTMLElement) error {
	start := time.Now()
	defer func() {
		p.lastProcessedTime = time.Now()
		atomic.AddInt64(&p.processedCount, 1)
		p.processingTime += time.Since(start)
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

	// Store content
	if err := p.storage.IndexDocument(context.Background(), p.indexName, content.ID, content); err != nil {
		atomic.AddInt64(&p.errorCount, 1)
		return fmt.Errorf("failed to store content: %w", err)
	}

	return nil
}

// GetMetrics returns the current processor metrics.
func (p *ContentProcessor) GetMetrics() *common.Metrics {
	return &common.Metrics{
		ProcessedCount:     atomic.LoadInt64(&p.processedCount),
		ErrorCount:         atomic.LoadInt64(&p.errorCount),
		LastProcessedTime:  p.lastProcessedTime,
		ProcessingDuration: p.processingTime,
	}
}

// ProcessJob processes a job and its items.
func (p *ContentProcessor) ProcessJob(ctx context.Context, job *common.Job) {
	start := time.Now()
	defer func() {
		p.processingTime += time.Since(start)
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		p.logger.Warn("Job processing cancelled",
			"job_id", job.ID,
			"error", ctx.Err(),
		)
		atomic.AddInt64(&p.errorCount, 1)
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

		atomic.AddInt64(&p.processedCount, 1)
	}
}
