// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/job"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// JobService implements the job.Service interface for the crawl module.
type JobService struct {
	logger           logger.Interface
	sources          sources.Interface
	crawler          crawler.Interface
	done             chan struct{}
	doneOnce         sync.Once    // Ensures done channel is only closed once
	activeJobs       atomic.Int32 // Use atomic.Int32 directly
	storage          storagetypes.Interface
	processorFactory crawler.ProcessorFactory
	sourceName       string
}

// JobServiceParams holds parameters for creating a new JobService.
type JobServiceParams struct {
	Logger           logger.Interface
	Sources          sources.Interface
	Crawler          crawler.Interface
	Done             chan struct{}
	Storage          storagetypes.Interface
	ProcessorFactory crawler.ProcessorFactory
	SourceName       string `name:"sourceName"`
}

// NewJobService creates a new JobService instance.
func NewJobService(p JobServiceParams) job.Service {
	return &JobService{
		logger:  p.Logger,
		sources: p.Sources,
		crawler: p.Crawler,
		done:    p.Done,
		// activeJobs is zero-initialized (no need to set it)
		storage:          p.Storage,
		processorFactory: p.ProcessorFactory,
		sourceName:       p.SourceName,
	}
}

// Start begins the job service.
func (s *JobService) Start(ctx context.Context) error {
	s.logger.Info("Starting job service")
	s.logger.Info("Starting crawl for source", "source", s.sourceName)

	// Start the crawler in a goroutine so it doesn't block
	go func() {
		// Start the crawler with the source name
		if err := s.crawler.Start(ctx, s.sourceName); err != nil {
			s.logger.Error("Crawler failed", "error", err)
		}
		// Wait for the crawler to complete all async operations
		// crawler.Start() returns immediately after starting async operations,
		// so we must wait for them to complete
		if err := s.crawler.Wait(); err != nil {
			s.logger.Error("Error waiting for crawler", "error", err)
		}
		// Signal completion when crawler finishes
		s.doneOnce.Do(func() {
			close(s.done)
		})
	}()

	return nil
}

// Stop implements the job.Service interface.
func (s *JobService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping crawl job")
	// Signal completion (safe to call multiple times)
	s.doneOnce.Do(func() {
		close(s.done)
	})
	return nil
}

// Status implements the job.Service interface.
func (s *JobService) Status(ctx context.Context) (content.JobStatus, error) {
	activeJobs := s.activeJobs.Load() // Use .Load() method
	state := content.JobStatusProcessing
	if activeJobs == 0 {
		state = content.JobStatusCompleted
	}
	return state, nil
}

// GetItems implements the job.Service interface.
func (s *JobService) GetItems(ctx context.Context, jobID string) ([]*content.Item, error) {
	s.logger.Info("Getting items for job", "jobID", jobID)
	// TODO: Implement item retrieval from storage
	return nil, nil
}

// UpdateItem implements the job.Service interface.
func (s *JobService) UpdateItem(ctx context.Context, item *content.Item) error {
	s.logger.Info("Updating item", "itemID", item.ID)
	// TODO: Implement item update in storage
	return nil
}

// UpdateJob implements the job.Service interface.
func (s *JobService) UpdateJob(ctx context.Context, jobObj *content.Job) error {
	s.logger.Info("Updating job", "jobID", jobObj.ID)
	// TODO: Implement job update in storage
	return nil
}
