package storage

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/mock"
)

// ErrMockTypeAssertion is returned when a type assertion fails in mock methods
var ErrMockTypeAssertion = errors.New("mock type assertion failed")

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

// Ensure MockStorage implements the Interface
var _ Interface = (*MockStorage)(nil)

// NewMockStorage creates a new instance of MockStorage
func NewMockStorage() *MockStorage {
	m := &MockStorage{}
	// Set up default expectations
	m.On("TestConnection", mock.Anything).Return(nil)
	m.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	m.On("Close").Return(nil)
	return m
}

// Close implements Interface
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// IndexDocument mocks base method
func (m *MockStorage) IndexDocument(ctx context.Context, indexName, docID string, document any) error {
	args := m.Called(ctx, indexName, docID, document)
	return args.Error(0)
}

// GetDocument mocks base method
func (m *MockStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// TestConnection mocks the connection test
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// BulkIndex mocks base method
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// Search mocks base method
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	result, ok := args.Get(0).([]any)
	if !ok {
		return nil, errors.New("could not convert result to []any")
	}
	return result, args.Error(1)
}

// CreateIndex mocks base method
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
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
	update map[string]any,
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
	query map[string]any,
	batchSize int,
) (<-chan map[string]any, error) {
	args := m.Called(ctx, index, query, batchSize)
	result, ok := args.Get(0).(<-chan map[string]any)
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// SearchArticles implements Interface
func (m *MockStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	args := m.Called(ctx, query, size)
	var articles []*models.Article

	// Check if the first argument is not nil before type assertion
	if args.Get(0) != nil {
		var ok bool
		articles, ok = args.Get(0).([]*models.Article)
		if !ok {
			return nil, ErrMockTypeAssertion // Return error if type assertion fails
		}
	}

	return articles, args.Error(1)
}

// IndexExists implements Interface
func (m *MockStorage) IndexExists(ctx context.Context, indexName string) (bool, error) {
	args := m.Called(ctx, indexName)
	return args.Bool(0), args.Error(1)
}

// IndexArticle implements Interface
func (m *MockStorage) IndexArticle(ctx context.Context, article *models.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

// Ping implements Interface
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// SearchDocuments implements Interface
func (m *MockStorage) SearchDocuments(
	ctx context.Context,
	index string,
	query string,
) ([]map[string]any, error) {
	args := m.Called(ctx, index, query)
	result, ok := args.Get(0).([]map[string]any)
	if !ok && args.Get(0) != nil {
		return nil, errors.New("invalid type assertion for SearchDocuments result")
	}
	return result, args.Error(1)
}

// IndexContent implements Interface
func (m *MockStorage) IndexContent(id string, content *models.Content) error {
	args := m.Called(id, content)
	return args.Error(0)
}

// GetContent implements Interface
func (m *MockStorage) GetContent(id string) (*models.Content, error) {
	args := m.Called(id)
	result, ok := args.Get(0).(*models.Content)
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// SearchContent implements Interface
func (m *MockStorage) SearchContent(query string) ([]*models.Content, error) {
	args := m.Called(query)
	result, ok := args.Get(0).([]*models.Content)
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// DeleteContent implements Interface
func (m *MockStorage) DeleteContent(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetMapping implements Interface
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	result, ok := args.Get(0).(map[string]any)
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// ListIndices implements Interface
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	result, ok := args.Get(0).([]string)
	if !ok && args.Get(0) != nil {
		return nil, ErrMockTypeAssertion
	}
	return result, args.Error(1)
}

// UpdateMapping implements Interface
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// GetIndexHealth implements Interface
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

// GetIndexDocCount implements Interface
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	result, ok := args.Get(0).(int64)
	if !ok && args.Get(0) != nil {
		return 0, errors.New("invalid type assertion for GetIndexDocCount result")
	}
	return result, args.Error(1)
}

// Aggregate implements Interface
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements Interface
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	result, ok := args.Get(0).(int64)
	if !ok && args.Get(0) != nil {
		return 0, errors.New("invalid type assertion for Count result")
	}
	return result, args.Error(1)
}
