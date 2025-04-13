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
	lastProcessed  time.Time
	duration       time.Duration
}

// NewCrawlerMetrics creates a new metrics tracker.
func NewCrawlerMetrics() CrawlerMetrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncrementProcessed increments the processed count.
func (m *Metrics) IncrementProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedCount++
	m.lastProcessed = time.Now()
	m.duration = time.Since(m.startTime)
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

// GetLastProcessedTime returns the time of the last processed item.
func (m *Metrics) GetLastProcessedTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastProcessed
}

// GetProcessingDuration returns the total processing duration.
func (m *Metrics) GetProcessingDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.duration
}

// Update updates the metrics with new values.
func (m *Metrics) Update(startTime time.Time, processed, errors int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startTime = startTime
	m.processedCount = processed
	m.errorCount = errors
	m.lastProcessed = time.Now()
	m.duration = time.Since(startTime)
}

// Reset resets all metrics to zero.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedCount = 0
	m.errorCount = 0
	m.startTime = time.Now()
	m.lastProcessed = time.Time{}
	m.duration = 0
}
