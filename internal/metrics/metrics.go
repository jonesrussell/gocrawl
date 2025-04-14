// Package metrics provides metrics collection and reporting functionality.
package metrics

import "time"

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
}

// NewMetrics creates a new Metrics instance with default values.
func NewMetrics() *Metrics {
	return &Metrics{
		ProcessedCount:     0,
		ErrorCount:         0,
		LastProcessedTime:  time.Time{},
		ProcessingDuration: 0,
	}
}
