package metrics

import (
	"sync"
	"time"
)

// Metrics tracks various statistics about the crawler's operation.
type Metrics struct {
	mu sync.RWMutex

	startTime           time.Time
	lastProcessedTime   time.Time
	processedCount      int64
	errorCount          int64
	currentSource       string
	successfulRequests  int64
	failedRequests      int64
	rateLimitedRequests int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// UpdateMetrics updates the metrics based on the processing result.
func (m *Metrics) UpdateMetrics(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processedCount++
	if !success {
		m.errorCount++
	}
	m.lastProcessedTime = time.Now()
}

// IncrementSuccessfulRequests increments the count of successful requests.
func (m *Metrics) IncrementSuccessfulRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.successfulRequests++
}

// IncrementFailedRequests increments the count of failed requests.
func (m *Metrics) IncrementFailedRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedRequests++
}

// IncrementRateLimitedRequests increments the count of rate-limited requests.
func (m *Metrics) IncrementRateLimitedRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimitedRequests++
}

// GetSuccessfulRequests returns the total number of successful requests.
func (m *Metrics) GetSuccessfulRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.successfulRequests
}

// GetFailedRequests returns the total number of failed requests.
func (m *Metrics) GetFailedRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failedRequests
}

// GetRateLimitedRequests returns the total number of rate-limited requests.
func (m *Metrics) GetRateLimitedRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rateLimitedRequests
}

// ResetMetrics resets all metrics to their initial state.
func (m *Metrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.startTime = time.Now()
	m.lastProcessedTime = time.Time{}
	m.processedCount = 0
	m.errorCount = 0
	m.currentSource = ""
	m.successfulRequests = 0
	m.failedRequests = 0
	m.rateLimitedRequests = 0
}

// GetStartTime returns the time when metrics collection started.
func (m *Metrics) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.startTime
}

// GetLastProcessedTime returns the time of the last processed item.
func (m *Metrics) GetLastProcessedTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastProcessedTime
}

// GetProcessedCount returns the total number of processed items.
func (m *Metrics) GetProcessedCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processedCount
}

// GetErrorCount returns the total number of errors encountered.
func (m *Metrics) GetErrorCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCount
}

// GetProcessingDuration returns the duration since metrics collection started.
func (m *Metrics) GetProcessingDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.startTime)
}

// SetCurrentSource sets the current source being processed.
func (m *Metrics) SetCurrentSource(source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentSource = source
}

// GetCurrentSource returns the current source being processed.
func (m *Metrics) GetCurrentSource() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentSource
}
