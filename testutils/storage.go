// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

var (
	// ErrInvalidMappingType is returned when the mapping type is invalid.
	ErrInvalidMappingType = errors.New("invalid mapping type")
)

// MockStorage is a mock implementation of the storage interface.
type MockStorage struct {
	mock.Mock
	logger logger.Interface
}

// NewMockStorage creates a new mock storage instance.
func NewMockStorage(log logger.Interface) types.Interface {
	return &MockStorage{
		logger: log,
	}
}

// IndexDocument indexes a document in Elasticsearch.
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument retrieves a document from Elasticsearch.
func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument deletes a document from Elasticsearch.
func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex performs bulk indexing operations.
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// BulkDelete performs bulk delete operations.
func (m *MockStorage) BulkDelete(ctx context.Context, index string, ids []string) error {
	args := m.Called(ctx, index, ids)
	return args.Error(0)
}

// CreateIndex creates a new index in Elasticsearch.
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex deletes an index from Elasticsearch.
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// IndexExists checks if an index exists in Elasticsearch.
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// GetMapping retrieves the mapping for an index.
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, ErrInvalidMappingType
	}
	return val, nil
}

// UpdateMapping updates the mapping for an index.
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// SearchDocuments performs a search query and decodes the result into the provided value
func (m *MockStorage) SearchDocuments(ctx context.Context, index string, query map[string]any, result any) error {
	args := m.Called(ctx, index, query, result)
	return args.Error(0)
}

// Search performs a search query in Elasticsearch.
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

// Count counts documents matching a query.
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

// Aggregate performs an aggregation query.
func (m *MockStorage) Aggregate(ctx context.Context, index string, query any) (any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0), args.Error(1)
}

// TestConnection tests the connection to Elasticsearch.
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Close closes the Elasticsearch client.
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SearchArticles searches for articles in Elasticsearch.
func (m *MockStorage) SearchArticles(ctx context.Context, query string, size int) ([]any, error) {
	args := m.Called(ctx, query, size)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

// GetIndexDocCount gets the document count for an index.
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

// GetIndexHealth gets the health status of an index.
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

// ListIndices lists all index in Elasticsearch.
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]string)
	if !ok {
		return nil, nil
	}
	return val, nil
}

// Ping pings the Elasticsearch cluster.
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// GetIndexManager returns the index manager for this storage
func (m *MockStorage) GetIndexManager() types.IndexManager {
	args := m.Called()
	if val, ok := args.Get(0).(types.IndexManager); ok {
		return val
	}
	return nil
}

// Ensure MockStorage implements storage.Interface
var _ types.Interface = (*MockStorage)(nil)
