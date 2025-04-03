// Package types provides storage types and interfaces.
package types

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockStorage implements Interface for testing.
type MockStorage struct {
	mock.Mock
}

// NewMockStorage creates a new mock storage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// IndexDocument implements Interface.
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument implements Interface.
func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument implements Interface.
func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex implements Interface.
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// CreateIndex implements Interface.
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex implements Interface.
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// ListIndices implements Interface.
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if result, ok := args.Get(0).([]string); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetMapping implements Interface.
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if result, ok := args.Get(0).(map[string]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateMapping implements Interface.
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// IndexExists implements Interface.
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// Search implements Interface.
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if result, ok := args.Get(0).([]any); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetIndexHealth implements Interface.
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

// GetIndexDocCount implements Interface.
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

// Ping implements Interface.
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestConnection implements Interface.
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Aggregate implements Interface.
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements Interface.
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

// Close implements Interface.
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}
