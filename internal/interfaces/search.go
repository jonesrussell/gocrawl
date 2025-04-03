// Package interfaces provides shared interfaces used across the application.
package interfaces

import "context"

// SearchManager defines the interface for search operations.
type SearchManager interface {
	// Search performs a search query.
	Search(ctx context.Context, index string, query any) ([]any, error)

	// Count returns the number of documents matching a query.
	Count(ctx context.Context, index string, query any) (int64, error)

	// Aggregate performs an aggregation query.
	Aggregate(ctx context.Context, index string, aggs any) (any, error)

	// Close closes any resources held by the search manager.
	Close() error
}
