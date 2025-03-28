package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
)

// MockIndexManager implements api.IndexManager for testing.
type MockIndexManager struct {
	api.IndexManager
}

// Index mocks the Index method.
func (m *MockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

// Close mocks the Close method.
func (m *MockIndexManager) Close() error {
	return nil
}

// NewMockIndexManager creates a new MockIndexManager instance.
func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}
