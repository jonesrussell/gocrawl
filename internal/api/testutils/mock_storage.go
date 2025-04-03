// Package testutils provides test utilities for the API package.
package testutils

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockStorage implements storage.Interface for testing.
type MockStorage struct {
	mock.Mock
}

// NewMockStorage creates a new mock storage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// Search implements storage.Interface.
func (m *MockStorage) Search(ctx context.Context, query, index string, size int) ([]any, error) {
	args := m.Called(ctx, query, index, size)
	if result, ok := args.Get(0).([]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// Count implements storage.Interface.
func (m *MockStorage) Count(ctx context.Context, query, index string) (int64, error) {
	args := m.Called(ctx, query, index)
	if result, ok := args.Get(0).(int64); ok {
		return result, args.Error(1)
	}
	return 0, args.Error(1)
}

// Close implements storage.Interface.
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockIndexManager implements interfaces.IndexManager for testing.
type MockIndexManager struct {
	mock.Mock
}

// NewMockIndexManager creates a new mock index manager instance.
func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}

// EnsureIndex implements interfaces.IndexManager.
func (m *MockIndexManager) EnsureIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// DeleteIndex implements interfaces.IndexManager.
func (m *MockIndexManager) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// IndexExists implements interfaces.IndexManager.
func (m *MockIndexManager) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// GetMapping implements interfaces.IndexManager.
func (m *MockIndexManager) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if result, ok := args.Get(0).(map[string]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateMapping implements interfaces.IndexManager.
func (m *MockIndexManager) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}
