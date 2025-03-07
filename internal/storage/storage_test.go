package storage_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ErrMockTypeAssertion is returned when a type assertion fails in mock methods
var ErrMockTypeAssertion = errors.New("mock type assertion failed")

func setupTestStorage(t *testing.T) (*storage.MockTransport, storage.Interface, *logger.MockLogger) {
	mockLogger := logger.NewMockLogger()
	mockTransport := &storage.MockTransport{}
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport,
	})
	require.NoError(t, err)

	result := storage.NewElasticsearchStorage(client, mockLogger, storage.DefaultOptions())
	require.NotNil(t, result.Storage)
	return mockTransport, result.Storage, mockLogger
}

func TestElasticsearchStorage_IndexDocument(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		id          string
		doc         interface{}
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:  "successful indexing",
			index: "test-index",
			id:    "1",
			doc:   map[string]interface{}{"title": "Test Document"},
			response: `{
				"_index": "test-index",
				"_id": "1",
				"result": "created"
			}`,
			statusCode:  201,
			expectError: false,
		},
		{
			name:        "indexing error",
			index:       "test-index",
			id:          "2",
			doc:         map[string]interface{}{"title": "Test Document"},
			response:    `{"error": {"type": "mapper_parsing_exception"}}`,
			statusCode:  400,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			if tt.expectError {
				mockLogger.On("Error", "Failed to index document", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Document indexed successfully", "index", tt.index, "docID", tt.id).Return()
			}

			err := store.IndexDocument(t.Context(), tt.index, tt.id, tt.doc)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_GetDocument(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		id          string
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:  "document found",
			index: "test-index",
			id:    "1",
			response: `{
				"_index": "test-index",
				"_id": "1",
				"_source": {
					"title": "Test Document"
				}
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "document not found",
			index:       "test-index",
			id:          "2",
			response:    `{"_index":"test-index","found":false}`,
			statusCode:  404,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			var result struct {
				Source struct {
					Title string `json:"title"`
				} `json:"_source"`
			}
			err := store.GetDocument(t.Context(), tt.index, tt.id, &result)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if !tt.expectError {
					require.NotNil(t, result.Source)
					assert.Equal(t, "Test Document", result.Source.Title)
				}
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_SearchArticles(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		query       string
		size        int
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:  "successful search",
			query: "test",
			size:  10,
			response: `{
				"hits": {
					"total": {"value": 1},
					"hits": [
						{
							"_source": {
								"title": "Test Article"
							}
						}
					]
				}
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "search error",
			query:       "test",
			size:        10,
			response:    `{"error": {"type": "search_phase_execution_exception"}}`,
			statusCode:  400,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			mockLogger.On("Debug", "Searching articles", "query", tt.query, "size", tt.size).Return()
			if tt.expectError {
				mockLogger.On("Error", "Failed to search documents", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Search completed", "query", tt.query, "results", int64(1)).Return()
			}

			articles, err := store.Search(t.Context(), tt.query, tt.size)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, articles)
			} else {
				require.NoError(t, err)
				require.NotNil(t, articles)
				require.Len(t, articles, 1)
				assert.Equal(t, "Test Article", articles[0].Title)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_BulkIndexArticles(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	articles := []*models.Article{
		{ID: "1", Title: "Test Article 1"},
		{ID: "2", Title: "Test Article 2"},
	}

	tests := []struct {
		name        string
		articles    []*models.Article
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:     "successful bulk indexing",
			articles: articles,
			response: `{
				"took": 30,
				"errors": false,
				"items": [
					{"index": {"_id": "1", "status": 201}},
					{"index": {"_id": "2", "status": 201}}
				]
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:     "bulk indexing with errors",
			articles: articles,
			response: `{
				"took": 30,
				"errors": true,
				"items": [
					{"index": {"_id": "1", "status": 400, "error": {"type": "mapper_parsing_exception"}}}
				]
			}`,
			statusCode:  400,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			mockLogger.On("Debug", "Bulk indexing articles", "count", len(tt.articles)).Return()
			if tt.expectError {
				mockLogger.On("Error", "Failed to bulk index articles", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Bulk indexed documents", "count", len(tt.articles)).Return()
			}

			err := store.BulkIndexArticles(t.Context(), tt.articles)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_TestConnection(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name: "successful connection",
			response: `{
				"name": "test-node",
				"cluster_name": "test-cluster",
				"version": {
					"number": "8.0.0"
				}
			}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "connection error",
			response:    `{"error": {"type": "connection_error"}}`,
			statusCode:  502,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			err := store.TestConnection(t.Context())
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_IndexExists(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name       string
		indexName  string
		statusCode int
		expected   bool
	}{
		{
			name:       "index exists",
			indexName:  "test-index",
			statusCode: http.StatusOK,
			expected:   true,
		},
		{
			name:       "index does not exist",
			indexName:  "nonexistent-index",
			statusCode: http.StatusNotFound,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.StatusCode = tt.statusCode

			exists, err := store.IndexExists(t.Context(), tt.indexName)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, exists)

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_CreateIndex(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		mapping     map[string]interface{}
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:  "successful index creation",
			index: "test-index",
			mapping: map[string]interface{}{
				"mappings": map[string]interface{}{
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type": "text",
						},
					},
				},
			},
			response:    `{"acknowledged": true}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:  "index creation error",
			index: "test-index",
			mapping: map[string]interface{}{
				"invalid": "mapping",
			},
			response:    `{"error": {"type": "invalid_mapping_exception"}}`,
			statusCode:  400,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to create index", "index", tt.index, "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Created index", "index", tt.index).Return()
			}

			err := store.CreateIndex(t.Context(), tt.index, tt.mapping)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_DeleteIndex(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful index deletion",
			index:       "test-index",
			response:    `{"acknowledged": true}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "index deletion error",
			index:       "nonexistent-index",
			response:    `{"error": {"type": "index_not_found_exception"}}`,
			statusCode:  404,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to delete index", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Deleted index", "index", tt.index).Return()
			}

			err := store.DeleteIndex(t.Context(), tt.index)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_DeleteDocument(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		docID       string
		response    string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful deletion",
			index:       "test-index",
			docID:       "1",
			response:    `{"_id": "1", "result": "deleted"}`,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "deletion error",
			index:       "test-index",
			docID:       "nonexistent",
			response:    `{"error": {"type": "document_missing_exception"}}`,
			statusCode:  404,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to delete document", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Deleted document", "index", tt.index, "docID", tt.docID).Return()
			}

			err := store.DeleteDocument(t.Context(), tt.index, tt.docID)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_GetMapping(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		response    string
		statusCode  int
		expectError bool
		expected    map[string]interface{}
	}{
		{
			name:  "successful mapping retrieval",
			index: "test-index",
			response: `{
				"test-index": {
					"mappings": {
						"properties": {
							"title": {"type": "text"}
						}
					}
				}
			}`,
			statusCode:  200,
			expectError: false,
			expected: map[string]interface{}{
				"test-index": map[string]interface{}{
					"mappings": map[string]interface{}{
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "text",
							},
						},
					},
				},
			},
		},
		{
			name:        "mapping retrieval error",
			index:       "nonexistent-index",
			response:    `{"error": {"type": "index_not_found_exception"}}`,
			statusCode:  404,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to get mapping", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Retrieved mapping", "index", tt.index).Return()
			}

			mapping, err := store.GetMapping(t.Context(), tt.index)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, mapping)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_ListIndices(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		response    string
		statusCode  int
		expectError bool
		expected    []string
	}{
		{
			name: "successful indices list",
			response: `[
				{"index": "index1"},
				{"index": "index2"}
			]`,
			statusCode:  200,
			expectError: false,
			expected:    []string{"index1", "index2"},
		},
		{
			name:        "list indices error",
			response:    `{"error": {"type": "cluster_block_exception"}}`,
			statusCode:  403,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to list indices", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Retrieved indices list").Return()
			}

			indices, err := store.ListIndices(t.Context())
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, indices)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_GetIndexHealth(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		response    string
		statusCode  int
		expectError bool
		expected    string
	}{
		{
			name:  "successful health check",
			index: "test-index",
			response: `{
				"status": "green"
			}`,
			statusCode:  200,
			expectError: false,
			expected:    "green",
		},
		{
			name:        "health check error",
			index:       "nonexistent-index",
			response:    `{"error": {"type": "index_not_found_exception"}}`,
			statusCode:  404,
			expectError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to get index health", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Retrieved index health", "index", tt.index, "health", tt.expected).Return()
			}

			health, err := store.GetIndexHealth(t.Context(), tt.index)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, health)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}

func TestElasticsearchStorage_GetIndexDocCount(t *testing.T) {
	mockTransport, store, mockLogger := setupTestStorage(t)

	tests := []struct {
		name        string
		index       string
		response    string
		statusCode  int
		expectError bool
		expected    int64
	}{
		{
			name:  "successful doc count",
			index: "test-index",
			response: `{
				"count": 42
			}`,
			statusCode:  200,
			expectError: false,
			expected:    42,
		},
		{
			name:        "doc count error",
			index:       "nonexistent-index",
			response:    `{"error": {"type": "index_not_found_exception"}}`,
			statusCode:  404,
			expectError: true,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransport.Response = tt.response
			mockTransport.StatusCode = tt.statusCode

			// Set up logger expectations
			if tt.expectError {
				mockLogger.On("Error", "Failed to get index document count", "error", mock.Anything).Return()
			} else {
				mockLogger.On("Info", "Retrieved index document count", "index", tt.index, "count", tt.expected).Return()
			}

			count, err := store.GetIndexDocCount(t.Context(), tt.index)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}
