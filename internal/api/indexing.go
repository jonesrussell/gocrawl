// Package api defines the interfaces for the application.
package api

import (
	"context"
)

// IndexManager defines the interface for managing Elasticsearch indices
type IndexManager interface {
	// EnsureIndex ensures that an index exists with the given name and mapping
	EnsureIndex(ctx context.Context, name string, mapping any) error

	// DeleteIndex deletes an index.
	DeleteIndex(ctx context.Context, name string) error

	// IndexExists checks if an index exists.
	IndexExists(ctx context.Context, name string) (bool, error)

	// UpdateMapping updates the mapping for an existing index
	UpdateMapping(ctx context.Context, name string, mapping any) error
}

// DocumentManager defines the interface for document operations.
type DocumentManager interface {
	// Index indexes a document with the given ID.
	Index(ctx context.Context, index string, id string, doc any) error

	// Update updates an existing document.
	Update(ctx context.Context, index string, id string, doc any) error

	// Delete deletes a document.
	Delete(ctx context.Context, index string, id string) error

	// Get retrieves a document by ID.
	Get(ctx context.Context, index string, id string) (any, error)
}
