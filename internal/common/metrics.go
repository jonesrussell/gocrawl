// Package common provides common types and interfaces used across the application.
package common

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
