// Package common provides common types and interfaces used across the application.
package common

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/content"
)

// JobService defines the interface for job operations.
type JobService interface {
	// Start starts the job service.
	Start(ctx context.Context) error
	// Stop stops the job service.
	Stop(ctx context.Context) error
	// Status returns the current status of the job service.
	Status(ctx context.Context) (content.JobStatus, error)
	// GetItems returns the items for a job.
	GetItems(ctx context.Context, jobID string) ([]*content.Item, error)
	// UpdateItem updates an item.
	UpdateItem(ctx context.Context, item *content.Item) error
	// UpdateJob updates a job.
	UpdateJob(ctx context.Context, job *content.Job) error
}
