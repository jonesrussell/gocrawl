// Package job provides core job service functionality.
package job

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Service provides base job service functionality.
type Service struct {
	logger     logger.Interface
	sources    sources.Interface
	crawler    crawler.Interface
	storage    types.Interface
	done       chan struct{}
	activeJobs *int32
	metrics    *metrics.Metrics
	validator  content.JobValidator
}

// ServiceParams holds parameters for creating a new Service.
type ServiceParams struct {
	Logger    logger.Interface
	Sources   sources.Interface
	Crawler   crawler.Interface
	Storage   types.Interface
	Done      chan struct{}
	Validator content.JobValidator
}

// NewService creates a new base job service.
func NewService(p ServiceParams) *Service {
	var jobs int32
	return &Service{
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
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting job service")
	return nil
}

// Stop stops the job service.
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping job service")
	close(s.done)
	return nil
}

// Status returns the current status of the job service.
func (s *Service) Status(ctx context.Context) (content.JobStatus, error) {
	activeJobs := atomic.LoadInt32(s.activeJobs)
	state := content.JobStatusProcessing
	if activeJobs == 0 {
		state = content.JobStatusCompleted
	}
	return state, nil
}

// GetItems returns the items for a job.
func (s *Service) GetItems(ctx context.Context, jobID string) ([]*content.Item, error) {
	s.logger.Info("Getting items for job", "jobID", jobID)
	return nil, fmt.Errorf("not implemented")
}

// UpdateItem updates an item.
func (s *Service) UpdateItem(ctx context.Context, item *content.Item) error {
	s.logger.Info("Updating item", "itemID", item.ID)
	return fmt.Errorf("not implemented")
}

// UpdateJob updates a job.
func (s *Service) UpdateJob(ctx context.Context, job *content.Job) error {
	s.logger.Info("Updating job", "jobID", job.ID)
	return fmt.Errorf("not implemented")
}

// ValidateJob validates a job.
func (s *Service) ValidateJob(job *content.Job) error {
	if s.validator == nil {
		return fmt.Errorf("no validator configured")
	}
	return s.validator.ValidateJob(job)
}

// IncrementActiveJobs increments the active job counter.
func (s *Service) IncrementActiveJobs() {
	atomic.AddInt32(s.activeJobs, 1)
}

// DecrementActiveJobs decrements the active job counter.
func (s *Service) DecrementActiveJobs() {
	atomic.AddInt32(s.activeJobs, -1)
}

// GetMetrics returns the current metrics.
func (s *Service) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// UpdateMetrics updates the metrics.
func (s *Service) UpdateMetrics(fn func(*metrics.Metrics)) {
	fn(s.metrics)
}

// GetLogger returns the logger.
func (s *Service) GetLogger() logger.Interface {
	return s.logger
}

// GetCrawler returns the crawler.
func (s *Service) GetCrawler() crawler.Interface {
	return s.crawler
}

// GetSources returns the sources.
func (s *Service) GetSources() sources.Interface {
	return s.sources
}

// GetStorage returns the storage.
func (s *Service) GetStorage() types.Interface {
	return s.storage
}

// IsDone returns true if the service is done.
func (s *Service) IsDone() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}
