// Package api defines the interfaces for the application.
package api

import "context"

// Search defines the interface for search operations.
type Search interface {
	// Search performs a search query.
	Search(ctx context.Context, query string) ([]any, error)
}
