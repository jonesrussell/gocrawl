// Package job provides the job scheduler implementation.
package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the job scheduler interface.
type Interface interface {
	// Start starts the job scheduler.
	Start(ctx context.Context) error
	// Stop stops the job scheduler.
	Stop() error
}

// Scheduler implements the job scheduler.
type Scheduler struct {
	logger   common.Logger
	sources  common.SourceManager
	storage  storagetypes.Interface
	crawler  crawler.Interface
	done     chan struct{}
	mu       sync.Mutex
	isActive bool
}

// New creates a new job scheduler.
func NewScheduler(logger common.Logger, sources common.SourceManager, storage storagetypes.Interface) *Scheduler {
	return &Scheduler{
		logger:  logger,
		sources: sources,
		storage: storage,
	}
}

// Start starts the job scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isActive {
		s.mu.Unlock()
		return nil
	}
	s.isActive = true
	s.mu.Unlock()

	s.logger.Info("Starting job scheduler")

	go func() {
		defer func() {
			s.mu.Lock()
			s.isActive = false
			s.mu.Unlock()
			s.logger.Info("Job scheduler stopped")
		}()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		// Run jobs immediately
		if err := s.runJobs(ctx); err != nil {
			s.logger.Error("Failed to run initial jobs", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Context cancelled, stopping job scheduler")
				return
			case <-s.done:
				s.logger.Info("Done signal received, stopping job scheduler")
				return
			case <-ticker.C:
				s.logger.Info("Running scheduled jobs")
				if err := s.runJobs(ctx); err != nil {
					s.logger.Error("Failed to run jobs", "error", err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the job scheduler.
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.isActive {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	s.logger.Info("Stopping job scheduler")
	close(s.done)
	return nil
}

// runJobs runs all configured jobs.
func (s *Scheduler) runJobs(ctx context.Context) error {
	sources, sourceErr := s.sources.GetSources()
	if sourceErr != nil {
		return fmt.Errorf("failed to get sources: %w", sourceErr)
	}
	s.logger.Info("Running jobs for sources", "count", len(sources))
	for _, source := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.logger.Info("Starting crawler for source", "source", source.Name)
			if err := s.crawler.Start(ctx, source.Name); err != nil {
				s.logger.Error("Failed to start crawler for source", "source", source.Name, "error", err)
			} else {
				s.logger.Info("Successfully started crawler for source", "source", source.Name)
			}
		}
	}
	return nil
}
