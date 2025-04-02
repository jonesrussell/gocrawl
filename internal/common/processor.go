// Package common provides common types and interfaces used across the application.
package common

import (
	"context"

	"github.com/gocolly/colly/v2"
)

// Processor defines the interface for processing jobs and their items.
type Processor interface {
	// ProcessJob processes a job and its items.
	ProcessJob(ctx context.Context, job *Job)
	// ProcessHTML processes HTML content from a source.
	ProcessHTML(e *colly.HTMLElement) error
	// GetMetrics returns the current processing metrics.
	GetMetrics() *Metrics
}
