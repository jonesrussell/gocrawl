// Package indexing defines interfaces and types for index management.
package indexing

import "context"

// Manager defines the interface for managing indices.
type Manager interface {
	// EnsureIndex ensures an index exists with the given mapping.
	EnsureIndex(ctx context.Context, name string, mapping interface{}) error

	// DeleteIndex deletes an index.
	DeleteIndex(ctx context.Context, name string) error

	// IndexExists checks if an index exists.
	IndexExists(ctx context.Context, name string) (bool, error)

	// UpdateMapping updates the mapping of an existing index.
	UpdateMapping(ctx context.Context, name string, mapping interface{}) error
}
