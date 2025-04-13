// Package types defines the core types and interfaces for storage operations.
package types

import (
	"context"
)

// IndexManager defines the interface for managing Elasticsearch indices.
type IndexManager interface {
	// EnsureIndex ensures an index exists with the given name and mapping
	EnsureIndex(ctx context.Context, name string, mapping any) error

	// DeleteIndex deletes an index with the given name
	DeleteIndex(ctx context.Context, name string) error

	// IndexExists checks if an index exists
	IndexExists(ctx context.Context, name string) (bool, error)

	// UpdateMapping updates the mapping for an index
	UpdateMapping(ctx context.Context, name string, mapping map[string]any) error

	// GetMapping gets the mapping for an index
	GetMapping(ctx context.Context, name string) (map[string]any, error)

	// EnsureArticleIndex ensures the article index exists
	EnsureArticleIndex(ctx context.Context, name string) error

	// EnsureContentIndex ensures the content index exists
	EnsureContentIndex(ctx context.Context, name string) error
}
