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

// IndexDocument implements types.Interface.
func (m *MockStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument implements types.Interface.
func (m *MockStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument implements types.Interface.
func (m *MockStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex implements types.Interface.
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// CreateIndex implements types.Interface.
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex implements types.Interface.
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// ListIndices implements types.Interface.
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if result, ok := args.Get(0).([]string); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetMapping implements types.Interface.
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if result, ok := args.Get(0).(map[string]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateMapping implements types.Interface.
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// IndexExists implements types.Interface.
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// Search implements types.Interface.
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if result, ok := args.Get(0).([]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetIndexHealth implements types.Interface.
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

// GetIndexDocCount implements types.Interface.
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(int64), args.Error(1)
}

// Ping implements types.Interface.
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestConnection implements types.Interface.
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Aggregate implements types.Interface.
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements types.Interface.
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

// Close implements types.Interface.
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
func (m *MockIndexManager) EnsureIndex(ctx context.Context, index string, mapping any) error {
	args := m.Called(ctx, index, mapping)
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
