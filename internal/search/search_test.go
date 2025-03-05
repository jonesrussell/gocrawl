package search_test

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	args := m.Called(ctx, query, size)
	return args.Get(0).([]*models.Article), args.Error(1)
}

func (m *mockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStorage) IndexDocument(ctx context.Context, index string, document interface{}) error {
	args := m.Called(ctx, index, document)
	return args.Error(0)
}

func (m *mockStorage) GetDocument(ctx context.Context, index string, id string) (interface{}, error) {
	args := m.Called(ctx, index, id)
	return args.Get(0), args.Error(1)
}

func (m *mockStorage) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStorage) BulkIndex(ctx context.Context, index string, documents []interface{}) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

func (m *mockStorage) Search(ctx context.Context, index string, query interface{}) (interface{}, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0), args.Error(1)
}

func (m *mockStorage) CreateIndex(ctx context.Context, index string, mapping interface{}) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *mockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

func (m *mockStorage) UpdateDocument(ctx context.Context, index string, id string, document interface{}) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *mockStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

func (m *mockStorage) ScrollSearch(ctx context.Context, index string, query interface{}, size int) (interface{}, error) {
	args := m.Called(ctx, index, query, size)
	return args.Get(0), args.Error(1)
}

func (m *mockStorage) BulkIndexArticles(ctx context.Context, articles []*models.Article) error {
	args := m.Called(ctx, articles)
	return args.Error(0)
}

func (m *mockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

func (m *mockStorage) IndexArticle(ctx context.Context, article *models.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m *mockStorage) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStorage) SearchDocuments(ctx context.Context, index string, query interface{}) ([]interface{}, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *mockStorage) IndexContent(ctx context.Context, content *models.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *mockStorage) GetContent(ctx context.Context, id string) (*models.Content, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Content), args.Error(1)
}

func (m *mockStorage) SearchContent(ctx context.Context, query string, size int) ([]*models.Content, error) {
	args := m.Called(ctx, query, size)
	return args.Get(0).([]*models.Content), args.Error(1)
}

func (m *mockStorage) DeleteContent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockStorage) GetMapping(ctx context.Context, index string) (interface{}, error) {
	args := m.Called(ctx, index)
	return args.Get(0), args.Error(1)
}

func (m *mockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockStorage) UpdateMapping(ctx context.Context, index string, mapping interface{}) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *mockStorage) GetIndexHealth(ctx context.Context, index string) (interface{}, error) {
	args := m.Called(ctx, index)
	return args.Get(0), args.Error(1)
}

func (m *mockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewSearchService(t *testing.T) {
	// Create mock dependencies
	mockLogger := logger.NewMockLogger()
	mockConfig := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
	}
	mockESClient := &elasticsearch.Client{}

	// Create service
	svc := search.NewSearchService(mockESClient, mockConfig, mockLogger)

	// Verify service fields
	assert.NotNil(t, svc)
	assert.Equal(t, mockESClient, svc.ESClient)
	assert.Equal(t, mockLogger, svc.Logger)
	assert.Equal(t, mockConfig, svc.Config)
	assert.NotNil(t, svc.Options)
}

func TestSearchContent(t *testing.T) {
	// Create mock dependencies
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "URL: http://example.com, Content: Test content").Return()

	mockConfig := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
	}
	mockESClient := &elasticsearch.Client{}

	// Create mock storage
	mockStorage := new(mockStorage)
	mockArticles := []*models.Article{
		{
			Source: "http://example.com",
			Body:   "Test content",
		},
	}
	mockStorage.On("SearchArticles", mock.Anything, "test query", 10).Return(mockArticles, nil)

	// Create service
	svc := search.NewSearchService(mockESClient, mockConfig, mockLogger)

	// Perform search
	results, err := svc.SearchContent(context.Background(), "test query", "", 10)

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "http://example.com", results[0].URL)
	assert.Equal(t, "Test content", results[0].Content)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
