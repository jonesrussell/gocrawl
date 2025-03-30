// Package processor provides content processing functionality for the application.
package processor

import (
	"sync/atomic"
	"time"
)

// MetricsCollector defines the interface for collecting metrics.
type MetricsCollector interface {
	// RecordProcessingTime records the time spent processing content.
	RecordProcessingTime(duration time.Duration)
	// RecordElementsProcessed records the number of elements processed.
	RecordElementsProcessed(count int64)
	// RecordError records an error occurrence.
	RecordError()
	// GetMetrics returns the current metrics.
	GetMetrics() *Metrics
	// Reset resets the metrics.
	Reset()
}

// DefaultMetricsCollector implements the MetricsCollector interface.
type DefaultMetricsCollector struct {
	metrics *Metrics
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		metrics: &Metrics{},
	}
}

// RecordProcessingTime records the time spent processing content.
func (m *DefaultMetricsCollector) RecordProcessingTime(duration time.Duration) {
	m.metrics.ProcessingDuration += duration
}

// RecordElementsProcessed records the number of elements processed.
func (m *DefaultMetricsCollector) RecordElementsProcessed(count int64) {
	atomic.AddInt64(&m.metrics.ProcessedCount, count)
}

// RecordError records an error occurrence.
func (m *DefaultMetricsCollector) RecordError() {
	atomic.AddInt64(&m.metrics.ErrorCount, 1)
}

// GetMetrics returns the current metrics.
func (m *DefaultMetricsCollector) GetMetrics() *Metrics {
	return m.metrics
}

// Reset resets the metrics.
func (m *DefaultMetricsCollector) Reset() {
	m.metrics = &Metrics{}
}
