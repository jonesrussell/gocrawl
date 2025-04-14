// Package metrics provides metrics collection and reporting functionality.
package metrics

import (
	"sync"
	"time"
)

// Metrics holds the processing metrics.
type Metrics struct {
	// ProcessedCount is the number of items processed.
	ProcessedCount int64
	// ErrorCount is the number of processing errors.
	ErrorCount int64
	// LastProcessedTime is the time of the last successful processing.
	LastProcessedTime time.Time
	// ProcessingDuration is the total time spent processing.
	ProcessingDuration time.Duration
	// StartTime is when the metrics collection began.
	StartTime time.Time
	// CurrentSource is the current source being processed.
	CurrentSource string
	// SuccessfulRequests is the number of successful HTTP requests.
	SuccessfulRequests int64
	// FailedRequests is the number of failed HTTP requests.
	FailedRequests int64
	// RateLimitedRequests is the number of rate-limited requests.
	RateLimitedRequests int64
	// mu protects concurrent access to metrics.
	mu sync.Mutex
}

// NewMetrics creates a new Metrics instance with default values.
func NewMetrics() *Metrics {
	return &Metrics{
		ProcessedCount:      0,
		ErrorCount:          0,
		LastProcessedTime:   time.Time{},
		ProcessingDuration:  0,
		StartTime:           time.Now(),
		CurrentSource:       "",
		SuccessfulRequests:  0,
		FailedRequests:      0,
		RateLimitedRequests: 0,
	}
}

// GetStartTime returns the time when metrics collection began.
func (m *Metrics) GetStartTime() time.Time {
	return m.StartTime
}

// UpdateMetrics updates the metrics based on processing success.
func (m *Metrics) UpdateMetrics(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ProcessedCount++
	if !success {
		m.ErrorCount++
	} else {
		m.LastProcessedTime = time.Now()
	}
}

// GetProcessedCount returns the number of items processed.
func (m *Metrics) GetProcessedCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ProcessedCount
}

// GetErrorCount returns the number of processing errors.
func (m *Metrics) GetErrorCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ErrorCount
}

// GetLastProcessedTime returns the time of the last successful processing.
func (m *Metrics) GetLastProcessedTime() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.LastProcessedTime
}

// GetProcessingDuration returns the total time spent processing.
func (m *Metrics) GetProcessingDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ProcessingDuration
}

// SetCurrentSource sets the current source being processed.
func (m *Metrics) SetCurrentSource(source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CurrentSource = source
}

// GetCurrentSource returns the current source being processed.
func (m *Metrics) GetCurrentSource() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CurrentSource
}

// ResetMetrics resets all metrics to their initial values.
func (m *Metrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ProcessedCount = 0
	m.ErrorCount = 0
	m.LastProcessedTime = time.Time{}
	m.ProcessingDuration = 0
	m.CurrentSource = ""
	m.SuccessfulRequests = 0
	m.FailedRequests = 0
	m.RateLimitedRequests = 0
}

// IncrementSuccessfulRequests increments the successful requests counter.
func (m *Metrics) IncrementSuccessfulRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SuccessfulRequests++
}

// GetSuccessfulRequests returns the number of successful requests.
func (m *Metrics) GetSuccessfulRequests() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.SuccessfulRequests
}

// IncrementFailedRequests increments the failed requests counter.
func (m *Metrics) IncrementFailedRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedRequests++
}

// GetFailedRequests returns the number of failed requests.
func (m *Metrics) GetFailedRequests() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.FailedRequests
}

// IncrementRateLimitedRequests increments the rate-limited requests counter.
func (m *Metrics) IncrementRateLimitedRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RateLimitedRequests++
}

// GetRateLimitedRequests returns the number of rate-limited requests.
func (m *Metrics) GetRateLimitedRequests() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.RateLimitedRequests
}
