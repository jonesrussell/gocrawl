package storage

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

// NewMockStorage creates a new instance of MockStorage
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

// IndexDocument mocks the indexing of a document
func (m *MockStorage) IndexDocument(ctx context.Context, indexName, docID string, document interface{}) error {
	args := m.Called(ctx, indexName, docID, document)
	return args.Error(0)
}

// TestConnection mocks the connection test
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// BulkIndex implements Storage
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []interface{}) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// Search implements Storage
func (m *MockStorage) Search(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// CreateIndex implements Storage
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex implements Storage
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// UpdateDocument implements Storage
func (m *MockStorage) UpdateDocument(ctx context.Context, index string, docID string, update map[string]interface{}) error {
	args := m.Called(ctx, index, docID, update)
	return args.Error(0)
}

// DeleteDocument implements Storage
func (m *MockStorage) DeleteDocument(ctx context.Context, index string, docID string) error {
	args := m.Called(ctx, index, docID)
	return args.Error(0)
}

// ScrollSearch implements Storage
func (m *MockStorage) ScrollSearch(ctx context.Context, index string, query map[string]interface{}, batchSize int) (<-chan map[string]interface{}, error) {
	args := m.Called(ctx, index, query, batchSize)
	return args.Get(0).(<-chan map[string]interface{}), args.Error(1)
}
