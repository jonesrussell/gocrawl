// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// JobService implements the common.JobService interface for the crawl module.
type JobService struct {
	logger           logger.Interface
	sources          *sources.Sources
	crawler          crawler.Interface
	done             chan struct{}
	config           config.Interface
	activeJobs       *int32
	storage          storagetypes.Interface
	processorFactory crawler.ProcessorFactory
	sourceName       string
}

// JobServiceParams holds parameters for creating a new JobService.
type JobServiceParams struct {
	Logger           logger.Interface
	Sources          *sources.Sources
	Crawler          crawler.Interface
	Done             chan struct{}
	Config           config.Interface
	Storage          storagetypes.Interface
	ProcessorFactory crawler.ProcessorFactory
	SourceName       string `name:"sourceName"`
}

// NewJobService creates a new JobService instance.
func NewJobService(p JobServiceParams) common.JobService {
	var jobs int32
	return &JobService{
		logger:           p.Logger,
		sources:          p.Sources,
		crawler:          p.Crawler,
		done:             p.Done,
		config:           p.Config,
		activeJobs:       &jobs,
		storage:          p.Storage,
		processorFactory: p.ProcessorFactory,
		sourceName:       p.SourceName,
	}
}

// Start begins the job service.
func (s *JobService) Start(ctx context.Context) error {
	s.logger.Info("Starting job service")
	s.logger.Info("Starting crawl for source", "source", s.sourceName)

	// Start the crawler with the source name
	if err := s.crawler.Start(ctx, s.sourceName); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Wait for the crawler to complete
	go func() {
		if err := s.crawler.Wait(); err != nil {
			s.logger.Error("Crawler failed", "error", err)
		}
		close(s.done)
	}()

	return nil
}

// Stop implements the common.JobService interface.
func (s *JobService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping crawl job")
	// Signal completion
	close(s.done)
	return nil
}

// Status implements the common.JobService interface.
func (s *JobService) Status(ctx context.Context) (jobtypes.JobStatus, error) {
	activeJobs := atomic.LoadInt32(s.activeJobs)
	state := jobtypes.JobStateRunning
	if activeJobs == 0 {
		state = jobtypes.JobStateCompleted
	}
	return jobtypes.JobStatus{
		State:     state,
		StartTime: time.Now(),
	}, nil
}

// GetItems implements the common.JobService interface.
func (s *JobService) GetItems(ctx context.Context, jobID string) ([]*jobtypes.Item, error) {
	s.logger.Info("Getting items for job", "jobID", jobID)
	// TODO: Implement item retrieval from storage
	return nil, nil
}

// UpdateItem implements the common.JobService interface.
func (s *JobService) UpdateItem(ctx context.Context, item *jobtypes.Item) error {
	s.logger.Info("Updating item", "itemID", item.ID)
	// TODO: Implement item update in storage
	return nil
}

// UpdateJob implements the common.JobService interface.
func (s *JobService) UpdateJob(ctx context.Context, job *jobtypes.Job) error {
	s.logger.Info("Updating job", "jobID", job.ID)
	// TODO: Implement job update in storage
	return nil
}
