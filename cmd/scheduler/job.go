// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// SchedulerService implements the common.JobService interface for the scheduler module.
type SchedulerService struct {
	logger           logger.Interface
	sources          sources.Interface
	crawler          crawler.Interface
	done             chan struct{}
	config           config.Interface
	activeJobs       *int32
	storage          types.Interface
	processorFactory crawler.ProcessorFactory
	items            map[string][]*content.Item
}

// NewSchedulerService creates a new SchedulerService instance.
func NewSchedulerService(
	logger logger.Interface,
	sources sources.Interface,
	crawler crawler.Interface,
	done chan struct{},
	config config.Interface,
	storage types.Interface,
	processorFactory crawler.ProcessorFactory,
) common.JobService {
	var jobs int32
	return &SchedulerService{
		logger:           logger,
		sources:          sources,
		crawler:          crawler,
		done:             done,
		config:           config,
		activeJobs:       &jobs,
		storage:          storage,
		processorFactory: processorFactory,
		items:            make(map[string][]*content.Item),
	}
}

// Start begins the scheduler service.
func (s *SchedulerService) Start(ctx context.Context) error {
	s.logger.Info("Starting scheduler service")

	// Check every minute
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Do initial check
	if err := s.checkAndRunJobs(ctx, time.Now()); err != nil {
		return fmt.Errorf("failed to run initial jobs: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping scheduler service")
			return nil
		case <-s.done:
			s.logger.Info("Done signal received, stopping scheduler service")
			return nil
		case t := <-ticker.C:
			if err := s.checkAndRunJobs(ctx, t); err != nil {
				s.logger.Error("Failed to run jobs", "error", err)
			}
		}
	}
}

// Stop stops the scheduler service.
func (s *SchedulerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping scheduler service")
	close(s.done)
	return nil
}

// GetItems returns the items collected by the scheduler service for a specific source.
func (s *SchedulerService) GetItems(ctx context.Context, sourceName string) ([]*content.Item, error) {
	items, ok := s.items[sourceName]
	if !ok {
		return nil, fmt.Errorf("no items found for source %s", sourceName)
	}
	return items, nil
}

// UpdateItem updates an item in the scheduler service.
func (s *SchedulerService) UpdateItem(ctx context.Context, item *content.Item) error {
	items, ok := s.items[item.Source]
	if !ok {
		return fmt.Errorf("no items found for source %s", item.Source)
	}
	for i, existingItem := range items {
		if existingItem.ID == item.ID {
			items[i] = item
			return nil
		}
	}
	return fmt.Errorf("item not found: %s", item.ID)
}

// Status returns the current status of the scheduler service.
func (s *SchedulerService) Status(ctx context.Context) (content.JobStatus, error) {
	state := content.JobStatusProcessing
	if atomic.LoadInt32(s.activeJobs) == 0 {
		state = content.JobStatusCompleted
	}
	return state, nil
}

// UpdateJob updates a job in the scheduler service.
func (s *SchedulerService) UpdateJob(ctx context.Context, job *content.Job) error {
	s.logger.Info("Updating job", "jobID", job.ID)
	// TODO: Implement job update in storage
	return nil
}

// checkAndRunJobs evaluates and executes scheduled jobs.
func (s *SchedulerService) checkAndRunJobs(ctx context.Context, now time.Time) error {
	if s.sources == nil {
		return errors.New("sources configuration is nil")
	}

	if s.crawler == nil {
		return errors.New("crawler instance is nil")
	}

	currentTime := now.Format("15:04")
	s.logger.Info("Checking jobs", "current_time", currentTime)

	// Execute crawl for each source
	sourcesList, err := s.sources.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	for i := range sourcesList {
		source := &sourcesList[i]
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				if crawlErr := s.executeCrawl(ctx, source); crawlErr != nil {
					s.logger.Error("Failed to execute crawl", "error", crawlErr)
					continue
				}
			}
		}
	}

	return nil
}

// executeCrawl performs the crawl operation for a single source.
func (s *SchedulerService) executeCrawl(ctx context.Context, source *sources.Config) error {
	atomic.AddInt32(s.activeJobs, 1)
	defer atomic.AddInt32(s.activeJobs, -1)

	// Start crawler
	if err := s.crawler.Start(ctx, source.URL); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Wait for completion
	if err := s.crawler.Wait(); err != nil {
		return fmt.Errorf("failed to wait for crawler: %w", err)
	}

	s.logger.Info("Crawl completed", "source", source.Name)
	return nil
}
