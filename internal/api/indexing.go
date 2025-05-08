// Package api defines the interfaces for the application.
package api

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// IndexManager is an alias for types.IndexManager
type IndexManager = types.IndexManager

// DocumentManager defines the interface for document management
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
