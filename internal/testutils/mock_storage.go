package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// MockStorage implements types.Interface for testing.
type MockStorage struct {
	types.Interface
}

// Store mocks the Store method.
func (m *MockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

// Close mocks the Close method.
func (m *MockStorage) Close() error {
	return nil
}

// NewMockStorage creates a new MockStorage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}
