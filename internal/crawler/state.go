// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"sync"
	"time"
)

// State implements the CrawlerState interface.
type State struct {
	mu            sync.RWMutex
	isRunning     bool
	startTime     time.Time
	currentSource string
	ctx           context.Context
	cancel        context.CancelFunc
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
