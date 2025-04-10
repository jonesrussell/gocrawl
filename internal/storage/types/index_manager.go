// Package types provides type definitions for the storage package.
package types

import "context"

// IndexManager defines the interface for managing Elasticsearch indices.
type IndexManager interface {
	// EnsureIndex creates an index if it doesn't exist
	EnsureIndex(ctx context.Context, index string) error
	// DeleteIndex deletes an index
	DeleteIndex(ctx context.Context, index string) error
	// ListIndices returns a list of all indices
	ListIndices(ctx context.Context) ([]string, error)
	// IndexExists checks if an index exists
	IndexExists(ctx context.Context, index string) (bool, error)
}
