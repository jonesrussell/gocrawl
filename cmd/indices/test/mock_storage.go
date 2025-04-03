// Package test provides test utilities for the indices command.
package test

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the storage interface
type MockStorage struct {
	mock.Mock
	types.Interface
}

// IndexDocument mocks the IndexDocument method
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument mocks the GetDocument method
func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument mocks the DeleteDocument method
func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex mocks the BulkIndex method
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// CreateIndex mocks the CreateIndex method
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex mocks the DeleteIndex method
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// ListIndices mocks the ListIndices method
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetMapping mocks the GetMapping method
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

// UpdateMapping mocks the UpdateMapping method
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// IndexExists mocks the IndexExists method
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(bool), args.Error(1)
}

// Search mocks the Search method
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

// GetIndexHealth mocks the GetIndexHealth method
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(string), args.Error(1)
}

// GetIndexDocCount mocks the GetIndexDocCount method
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(int64), args.Error(1)
}

// Ping mocks the Ping method
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestConnection mocks the TestConnection method
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Aggregate mocks the Aggregate method
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

// Count mocks the Count method
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

// Close mocks the Close method
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}
