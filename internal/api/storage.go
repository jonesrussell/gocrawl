// Package api defines the interfaces for the application.
package api

import (
	"context"
)

// Storage defines the storage operations.
type Storage interface {
	// TestConnection tests the connection to the storage backend.
	TestConnection(ctx context.Context) error
}
