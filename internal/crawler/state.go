// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"sync"
	"time"
)

// State implements the CrawlerState interface.
type State struct {
	mu             sync.RWMutex
	isRunning      bool
	startTime      time.Time
	currentSource  string
	ctx            context.Context
	cancel         context.CancelFunc
	processedCount int64
	errorCount     int64
}

// NewState creates a new crawler state.
func NewState() CrawlerState {
	return &State{}
}

// IsRunning returns whether the crawler is running.
func (s *State) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// StartTime returns when the crawler started.
func (s *State) StartTime() time.Time {
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
	s.isRunning = false
	s.currentSource = ""
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.ctx = nil
}

// Update updates the state with new values.
func (s *State) Update(startTime time.Time, processed int64, errors int64) {
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
