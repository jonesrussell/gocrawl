// Package common provides common types and interfaces used across the application.
package common

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/common/jobtypes"
)

// JobService defines the interface for job operations.
type JobService interface {
	// GetItems returns the items for a job.
	GetItems(ctx context.Context, jobID string) ([]*jobtypes.Item, error)
	// UpdateItem updates an item.
	UpdateItem(ctx context.Context, item *jobtypes.Item) error
	// UpdateJob updates a job.
	UpdateJob(ctx context.Context, job *jobtypes.Job) error
}
