// Package types defines the core types and interfaces for storage operations.
package types

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/interfaces"
)

// Interface defines the interface for storage operations.
type Interface interface {
	// GetIndexManager returns the index manager for this storage
	GetIndexManager() interfaces.IndexManager

	// Document operations
	IndexDocument(ctx context.Context, index string, id string, document any) error
	GetDocument(ctx context.Context, index string, id string, document any) error
	DeleteDocument(ctx context.Context, index string, id string) error
	SearchDocuments(ctx context.Context, index string, query map[string]any, result any) error

	// Search operations
	Search(ctx context.Context, index string, query any) ([]any, error)
	Count(ctx context.Context, index string, query any) (int64, error)
	Aggregate(ctx context.Context, index string, aggs any) (any, error)

	// Index operations
	CreateIndex(ctx context.Context, index string, mapping map[string]any) error
	DeleteIndex(ctx context.Context, index string) error
	IndexExists(ctx context.Context, index string) (bool, error)
	ListIndices(ctx context.Context) ([]string, error)
	GetMapping(ctx context.Context, index string) (map[string]any, error)
	UpdateMapping(ctx context.Context, index string, mapping map[string]any) error
	GetIndexHealth(ctx context.Context, index string) (string, error)
	GetIndexDocCount(ctx context.Context, index string) (int64, error)

	// Connection operations
	TestConnection(ctx context.Context) error
	Close() error
}
