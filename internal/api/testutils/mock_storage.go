// Package testutils provides test utilities for the API package.
package testutils

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/interfaces"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
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

// IndexDocument implements storage.Interface.
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument implements storage.Interface.
func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument implements storage.Interface.
func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex implements storage.Interface.
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// CreateIndex implements storage.Interface.
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex implements storage.Interface.
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// ListIndices implements storage.Interface.
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).([]string)
	if !ok {
		return nil, errors.New("invalid indices result type")
	}
	return result, nil
}

// GetMapping implements storage.Interface.
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, errors.New("invalid mapping result type")
	}
	return result, nil
}

// UpdateMapping implements storage.Interface.
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// IndexExists implements storage.Interface.
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// SearchDocuments performs a search query and decodes the result into the provided value
func (m *MockStorage) SearchDocuments(ctx context.Context, index string, query map[string]any, result any) error {
	args := m.Called(ctx, index, query, result)
	return args.Error(0)
}

// Search performs a search query
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]any), args.Error(1)
}

// Count returns the number of documents matching the query
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

// Aggregate performs an aggregation query
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// GetIndexHealth implements storage.Interface.
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return "", err
	}
	result, ok := args.Get(0).(string)
	if !ok {
		return "", errors.New("invalid health result type")
	}
	return result, nil
}

// GetIndexDocCount implements storage.Interface.
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	result, ok := args.Get(0).(int64)
	if !ok {
		return 0, errors.New("invalid doc count result type")
	}
	return result, nil
}

// Ping implements storage.Interface.
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestConnection implements storage.Interface.
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
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
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, errors.New("invalid mapping result type")
	}
	return result, nil
}

// UpdateMapping implements interfaces.IndexManager.
func (m *MockIndexManager) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// EnsureArticleIndex implements interfaces.IndexManager.
func (m *MockIndexManager) EnsureArticleIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// EnsureContentIndex implements interfaces.IndexManager.
func (m *MockIndexManager) EnsureContentIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// GetIndexManager implements storage.Interface.
func (m *MockStorage) GetIndexManager() interfaces.IndexManager {
	args := m.Called()
	if val, ok := args.Get(0).(interfaces.IndexManager); ok {
		return val
	}
	return nil
}

// Ensure MockStorage implements storage.Interface
var _ storagetypes.Interface = (*MockStorage)(nil)
