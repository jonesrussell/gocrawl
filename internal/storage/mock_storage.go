package storage

import (
	"context"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct{}

// NewMockStorage creates a new instance of MockStorage
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// IndexDocument mocks the indexing of a document
func (m *MockStorage) IndexDocument(_ context.Context, _ string, _ string, _ interface{}) error {
	// Simulate successful indexing
	return nil
}

// TestConnection mocks the connection test
func (m *MockStorage) TestConnection(_ context.Context) error {
	// Simulate successful connection
	return nil
}
