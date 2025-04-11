// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"sync"
	"time"
)

// Metrics implements the CrawlerMetrics interface.
type Metrics struct {
	mu             sync.RWMutex
	processedCount int64
	errorCount     int64
	startTime      time.Time
}

// NewMetrics creates a new metrics tracker.
func NewMetrics() CrawlerMetrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncrementProcessed increments the processed count.
func (m *Metrics) IncrementProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedCount++
}

// IncrementError increments the error count.
func (m *Metrics) IncrementError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount++
}

// GetProcessedCount returns the number of processed items.
func (m *Metrics) GetProcessedCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processedCount
}

// GetErrorCount returns the number of errors.
func (m *Metrics) GetErrorCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCount
}

// GetStartTime returns when tracking started.
func (m *Metrics) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.startTime
}

// Reset resets all metrics to zero.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedCount = 0
	m.errorCount = 0
	m.startTime = time.Now()
}
