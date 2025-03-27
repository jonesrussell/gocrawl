package types

import (
	"context"
)

// Interface defines the storage operations
type Interface interface {
	// Document operations
	IndexDocument(ctx context.Context, index string, id string, document any) error
	GetDocument(ctx context.Context, index string, id string, document any) error
	DeleteDocument(ctx context.Context, index string, id string) error

	// Bulk operations
	BulkIndex(ctx context.Context, index string, documents []any) error

	// Index management
	CreateIndex(ctx context.Context, index string, mapping map[string]any) error
	DeleteIndex(ctx context.Context, index string) error
	ListIndices(ctx context.Context) ([]string, error)
	GetMapping(ctx context.Context, index string) (map[string]any, error)
	UpdateMapping(ctx context.Context, index string, mapping map[string]any) error
	IndexExists(ctx context.Context, index string) (bool, error)

	// Search operations
	Search(ctx context.Context, index string, query any) ([]any, error)

	// Index health and stats
	GetIndexHealth(ctx context.Context, index string) (string, error)
	GetIndexDocCount(ctx context.Context, index string) (int64, error)

	// Common operations
	Ping(ctx context.Context) error
	TestConnection(ctx context.Context) error

	// New operations
	Aggregate(ctx context.Context, index string, aggs any) (any, error)

	// Count operation
	Count(ctx context.Context, index string, query any) (int64, error)
}
