// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/job"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// SchedulerService implements the job.Service interface for the scheduler module.
type SchedulerService struct {
	logger           logger.Interface
	sources          sources.Interface
	crawler          crawler.Interface
	done             chan struct{}
	doneOnce         sync.Once // Ensures done channel is only closed once
	config           config.Interface
	activeJobs       atomic.Int32 // Use atomic.Int32 directly
	storage          types.Interface
	processorFactory crawler.ProcessorFactory
	items            map[string][]*content.Item
}

// NewSchedulerService creates a new SchedulerService instance.
func NewSchedulerService(
	log logger.Interface,
	sourcesManager sources.Interface,
	crawlerInstance crawler.Interface,
	done chan struct{},
	cfg config.Interface,
	storage types.Interface,
	processorFactory crawler.ProcessorFactory,
) job.Service {
	return &SchedulerService{
		logger:  log,
		sources: sourcesManager,
		crawler: crawlerInstance,
		done:    done,
		config:  cfg,
		// activeJobs is zero-initialized
		storage:          storage,
		processorFactory: processorFactory,
		items:            make(map[string][]*content.Item),
	}
}

// Start begins the scheduler service.
func (s *SchedulerService) Start(ctx context.Context) error {
	s.logger.Info("Starting scheduler service")

	// Start the scheduler loop in a goroutine so it doesn't block
	go func() {
		// Check every minute
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		// Do initial check
		if err := s.checkAndRunJobs(ctx, time.Now()); err != nil {
			s.logger.Error("Failed to run initial jobs", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Context cancelled, stopping scheduler service")
				return
			case <-s.done:
				s.logger.Info("Done signal received, stopping scheduler service")
				return
			case t := <-ticker.C:
				if err := s.checkAndRunJobs(ctx, t); err != nil {
					s.logger.Error("Failed to run jobs", "error", err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the scheduler service.
func (s *SchedulerService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping scheduler service")
	// Signal completion (safe to call multiple times)
	s.doneOnce.Do(func() {
		close(s.done)
	})
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
	if s.activeJobs.Load() == 0 { // Use .Load() method
		state = content.JobStatusCompleted
	}
	return state, nil
}

// UpdateJob updates a job in the scheduler service.
func (s *SchedulerService) UpdateJob(ctx context.Context, jobObj *content.Job) error {
	s.logger.Info("Updating job", "jobID", jobObj.ID)
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
	s.activeJobs.Add(1)        // Use .Add() method
	defer s.activeJobs.Add(-1) // Use .Add(-1) instead of atomic.AddInt32

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
