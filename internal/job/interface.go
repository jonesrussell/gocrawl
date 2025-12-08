// Package job provides core job service functionality.
package job

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/content"
)

// Service defines the interface for job operations.
// This interface combines lifecycle management (Start, Stop, Status) with data operations
// (GetItems, UpdateItem, UpdateJob). Some implementations may not fully implement all data
// operations yet - in such cases, they should return nil/empty values rather than errors.
type Service interface {
	// Start starts the job service. This is a required method for all implementations.
	Start(ctx context.Context) error
	// Stop stops the job service. This is a required method for all implementations.
	Stop(ctx context.Context) error
	// Status returns the current status of the job service. This is a required method for all implementations.
	Status(ctx context.Context) (content.JobStatus, error)
	// GetItems returns the items for a job.
	// Note: Some implementations may return nil/empty slice if item retrieval is not yet implemented.
	GetItems(ctx context.Context, jobID string) ([]*content.Item, error)
	// UpdateItem updates an item.
	// Note: Some implementations may be stubbed and return nil if item updates are not yet implemented.
	UpdateItem(ctx context.Context, item *content.Item) error
	// UpdateJob updates a job.
	// Note: Some implementations may be stubbed and return nil if job updates are not yet implemented.
	UpdateJob(ctx context.Context, job *content.Job) error
}
