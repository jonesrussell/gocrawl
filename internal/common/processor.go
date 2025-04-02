// Package common provides common types and interfaces used across the application.
package common

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
)

// Processor defines the interface for processing jobs and their items.
type Processor interface {
	// ProcessJob processes a job and its items.
	ProcessJob(ctx context.Context, job *Job)
	// ProcessHTML processes HTML content from a source.
	ProcessHTML(e *colly.HTMLElement) error
	// Process processes the given data.
	Process(ctx context.Context, data any) error
	// Start starts the processor.
	Start(ctx context.Context) error
	// Stop stops the processor.
	Stop(ctx context.Context) error
	// GetMetrics returns the current processing metrics.
	GetMetrics() *Metrics
}

// NoopProcessor is a no-op implementation of Processor.
type NoopProcessor struct{}

// NewNoopProcessor creates a new no-op processor.
func NewNoopProcessor() *NoopProcessor {
	return &NoopProcessor{}
}

// ProcessJob implements Processor.ProcessJob.
func (p *NoopProcessor) ProcessJob(ctx context.Context, job *Job) {
}

// ProcessHTML implements Processor.ProcessHTML.
func (p *NoopProcessor) ProcessHTML(e *colly.HTMLElement) error {
	return nil
}

// Process implements Processor.Process.
func (p *NoopProcessor) Process(ctx context.Context, data any) error {
	return nil
}

// Start implements Processor.Start.
func (p *NoopProcessor) Start(ctx context.Context) error {
	return nil
}

// Stop implements Processor.Stop.
func (p *NoopProcessor) Stop(ctx context.Context) error {
	return nil
}

// GetMetrics implements Processor.GetMetrics.
func (p *NoopProcessor) GetMetrics() *Metrics {
	return &Metrics{
		ProcessedCount:     0,
		ErrorCount:         0,
		LastProcessedTime:  time.Time{},
		ProcessingDuration: 0,
	}
}
