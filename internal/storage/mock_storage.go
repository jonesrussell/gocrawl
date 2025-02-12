package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/mock"
)

// ErrMockTypeAssertion is returned when a type assertion fails in mock methods
var ErrMockTypeAssertion = errors.New("mock type assertion failed")

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

// NewMockStorage creates a new instance of MockStorage
func NewMockStorage() *MockStorage {
	m := &MockStorage{}
	// Set up default expectations
	m.On("TestConnection", mock.Anything).Return(nil)
	m.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	m.On("BulkIndexArticles", mock.Anything, mock.Anything).Return(nil)
	return m
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
func (m *MockStorage) Search(
	ctx context.Context,
	index string,
	query map[string]interface{},
) ([]map[string]interface{}, error) {
	args := m.Called(ctx, index, query)
	result, ok := args.Get(0).([]map[string]interface{})
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
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
func (m *MockStorage) UpdateDocument(
	ctx context.Context,
	index string,
	docID string,
	update map[string]interface{},
) error {
	args := m.Called(ctx, index, docID, update)
	return args.Error(0)
}

// DeleteDocument implements Storage
func (m *MockStorage) DeleteDocument(ctx context.Context, index string, docID string) error {
	args := m.Called(ctx, index, docID)
	return args.Error(0)
}

// ScrollSearch implements Storage
func (m *MockStorage) ScrollSearch(
	ctx context.Context,
	index string,
	query map[string]interface{},
	batchSize int,
) (<-chan map[string]interface{}, error) {
	args := m.Called(ctx, index, query, batchSize)
	result, ok := args.Get(0).(<-chan map[string]interface{})
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// BulkIndexArticles implements Storage
func (m *MockStorage) BulkIndexArticles(ctx context.Context, articles []*models.Article) error {
	args := m.Called(ctx, articles)
	return args.Error(0)
}

// SearchArticles implements Storage
func (m *MockStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	args := m.Called(ctx, query, size)
	var articles []*models.Article
	if args.Get(0) != nil {
		articles = args.Get(0).([]*models.Article)
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("error getting articles: %w", err)
		}
	}
	return articles, nil
}

// IndexExists implements Storage
func (m *MockStorage) IndexExists(ctx context.Context, indexName string) (bool, error) {
	args := m.Called(ctx, indexName)
	return args.Bool(0), args.Error(1)
}
