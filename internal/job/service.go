// Package job provides core job service functionality.
package job

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// BaseService provides base job service functionality.
type BaseService struct {
	logger     logger.Interface
	sources    sources.Interface
	crawler    crawler.Interface
	storage    types.Interface
	done       chan struct{}
	activeJobs *int32
	metrics    *metrics.Metrics
	validator  content.JobValidator
}

// ServiceParams holds parameters for creating a new BaseService.
type ServiceParams struct {
	Logger    logger.Interface
	Sources   sources.Interface
	Crawler   crawler.Interface
	Storage   types.Interface
	Done      chan struct{}
	Validator content.JobValidator
}

// NewService creates a new base job service.
func NewService(p ServiceParams) Service {
	var jobs int32
	return &BaseService{
		logger:     p.Logger,
		sources:    p.Sources,
		crawler:    p.Crawler,
		storage:    p.Storage,
		done:       p.Done,
		activeJobs: &jobs,
		metrics:    metrics.NewMetrics(),
		validator:  p.Validator,
	}
}

// Start starts the job service.
func (s *BaseService) Start(ctx context.Context) error {
	s.logger.Info("Starting job service")
	return nil
}

// Stop stops the job service.
func (s *BaseService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping job service")
	close(s.done)
	return nil
}

// Status returns the current status of the job service.
func (s *BaseService) Status(ctx context.Context) (content.JobStatus, error) {
	activeJobs := atomic.LoadInt32(s.activeJobs)
	state := content.JobStatusProcessing
	if activeJobs == 0 {
		state = content.JobStatusCompleted
	}
	return state, nil
}

// GetItems returns the items for a job.
func (s *BaseService) GetItems(ctx context.Context, jobID string) ([]*content.Item, error) {
	s.logger.Info("Getting items for job", "jobID", jobID)
	return nil, errors.New("not implemented")
}

// UpdateItem updates an item.
func (s *BaseService) UpdateItem(ctx context.Context, item *content.Item) error {
	s.logger.Info("Updating item", "itemID", item.ID)
	return errors.New("not implemented")
}

// UpdateJob updates a job.
func (s *BaseService) UpdateJob(ctx context.Context, job *content.Job) error {
	s.logger.Info("Updating job", "jobID", job.ID)
	return errors.New("not implemented")
}

// ValidateJob validates a job.
func (s *BaseService) ValidateJob(job *content.Job) error {
	if s.validator == nil {
		return errors.New("no validator configured")
	}
	return s.validator.ValidateJob(job)
}

// IncrementActiveJobs increments the active job counter.
func (s *BaseService) IncrementActiveJobs() {
	atomic.AddInt32(s.activeJobs, 1)
}

// DecrementActiveJobs decrements the active job counter.
func (s *BaseService) DecrementActiveJobs() {
	atomic.AddInt32(s.activeJobs, -1)
}

// GetMetrics returns the current metrics.
func (s *BaseService) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// UpdateMetrics updates the metrics.
func (s *BaseService) UpdateMetrics(fn func(*metrics.Metrics)) {
	fn(s.metrics)
}

// GetLogger returns the logger.
func (s *BaseService) GetLogger() logger.Interface {
	return s.logger
}

// GetCrawler returns the crawler.
func (s *BaseService) GetCrawler() crawler.Interface {
	return s.crawler
}

// GetSources returns the sources.
func (s *BaseService) GetSources() sources.Interface {
	return s.sources
}

// GetStorage returns the storage.
func (s *BaseService) GetStorage() types.Interface {
	return s.storage
}

// IsDone returns true if the service is done.
func (s *BaseService) IsDone() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

// Ensure BaseService implements Service interface
var _ Service = (*BaseService)(nil)
