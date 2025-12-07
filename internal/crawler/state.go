// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// State implements the CrawlerState and CrawlerMetrics interfaces.
type State struct {
	mu                sync.RWMutex
	isRunning         bool
	startTime         time.Time
	currentSource     string
	ctx               context.Context
	cancel            context.CancelFunc
	processedCount    int64
	errorCount        int64
	lastProcessedTime time.Time
	logger            logger.Interface
}

// NewState creates a new crawler state.
func NewState(log logger.Interface) *State {
	return &State{
		logger: log,
	}
}

// IsRunning returns whether the crawler is running.
func (s *State) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetStartTime returns when the crawler started.
func (s *State) GetStartTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startTime
}

// CurrentSource returns the current source being crawled.
func (s *State) CurrentSource() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSource
}

// Context returns the crawler's context.
func (s *State) Context() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ctx
}

// Cancel cancels the crawler's context.
func (s *State) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
}

// Start initializes the crawler state.
func (s *State) Start(ctx context.Context, sourceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = true
	s.startTime = time.Now()
	s.currentSource = sourceName
	s.ctx, s.cancel = context.WithCancel(ctx)
}

// Stop cleans up the crawler state.
func (s *State) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel any existing context
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	// Reset state
	s.isRunning = false
	s.currentSource = ""
	s.ctx = nil

	// Log metrics
	s.logger.Info("Crawler stopped",
		"processed", s.processedCount,
		"errors", s.errorCount,
		"duration", time.Since(s.startTime))
}

// Update updates the metrics with new values.
func (s *State) Update(startTime time.Time, processed, errors int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startTime = startTime
	s.processedCount = processed
	s.errorCount = errors
}

// Reset resets the state to its initial values.
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = false
	s.startTime = time.Time{}
	s.currentSource = ""
	s.processedCount = 0
	s.errorCount = 0
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.ctx = nil
}

// GetProcessedCount returns the number of processed items.
func (s *State) GetProcessedCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.processedCount
}

// GetErrorCount returns the number of errors.
func (s *State) GetErrorCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorCount
}

// GetLastProcessedTime returns the time of the last processed item.
func (s *State) GetLastProcessedTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastProcessedTime
}

// GetProcessingDuration returns the total processing duration.
func (s *State) GetProcessingDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isRunning {
		return time.Duration(0)
	}
	return time.Since(s.startTime)
}

// IncrementProcessed increments the processed count.
func (s *State) IncrementProcessed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processedCount++
	s.lastProcessedTime = time.Now()
}

// IncrementError increments the error count.
func (s *State) IncrementError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorCount++
}
