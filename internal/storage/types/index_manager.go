// Package types defines the core types and interfaces for storage operations.
package types

import "context"

// IndexManager defines the interface for managing Elasticsearch index.
type IndexManager interface {
	// EnsureIndex ensures that an index exists with the specified mapping.
	EnsureIndex(ctx context.Context, name string, mapping any) error
	// DeleteIndex deletes an index.
	DeleteIndex(ctx context.Context, name string) error
	// IndexExists checks if an index exists.
	IndexExists(ctx context.Context, name string) (bool, error)
	// GetMapping gets the mapping for an index.
	GetMapping(ctx context.Context, name string) (map[string]any, error)
	// UpdateMapping updates the mapping for an index.
	UpdateMapping(ctx context.Context, name string, mapping map[string]any) error
	// EnsureArticleIndex ensures that an article index exists with the appropriate mapping.
	EnsureArticleIndex(ctx context.Context, name string) error
	// EnsurePageIndex ensures that a page index exists with the appropriate mapping.
	EnsurePageIndex(ctx context.Context, name string) error
}
