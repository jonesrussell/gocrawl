package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSearchManager struct {
	mock.Mock
}

func (m *mockSearchManager) Search(ctx context.Context, index string, query interface{}) ([]interface{}, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query interface{}) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSearchManager) Aggregate(ctx context.Context, index string, aggs interface{}) (interface{}, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	args := m.Called()
	return args.Get(0).(*config.CrawlerConfig)
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	return args.Get(0).(*config.ElasticsearchConfig)
}

func (m *mockConfig) GetLogConfig() *config.LogConfig {
	args := m.Called()
	return args.Get(0).(*config.LogConfig)
}

func (m *mockConfig) GetAppConfig() *config.AppConfig {
	args := m.Called()
	return args.Get(0).(*config.AppConfig)
}

func (m *mockConfig) GetSources() []config.Source {
	args := m.Called()
	return args.Get(0).([]config.Source)
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	args := m.Called()
	return args.Get(0).(*config.ServerConfig)
}

func TestStartHTTPServer(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		index          string
		size           int
		mockResults    []interface{}
		mockCount      int64
		expectedStatus int
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "test query",
			index: "articles",
			size:  10,
			mockResults: []interface{}{
				map[string]interface{}{
					"id":    "1",
					"title": "Test Article",
				},
			},
			mockCount:      1,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty results",
			query:          "nonexistent",
			index:          "articles",
			size:           10,
			mockResults:    []interface{}{},
			mockCount:      0,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger
			mockLogger := logger.NewMockLogger()
			mockLogger.On("Info", "StartHTTPServer function called").Return()

			// Create mock search manager
			mockSearch := new(mockSearchManager)

			// Create mock config
			mockCfg := new(mockConfig)
			mockCfg.On("GetServerConfig").Return(&config.ServerConfig{
				Address:      ":8080",
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			})

			// Set up expected query
			expectedQuery := map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"content": tt.query,
					},
				},
				"size": tt.size,
			}

			// Set up mock expectations
			mockSearch.On("Search", mock.Anything, tt.index, expectedQuery).Return(tt.mockResults, nil)
			mockSearch.On("Count", mock.Anything, tt.index, expectedQuery).Return(tt.mockCount, nil)

			// Start the server
			server, err := api.StartHTTPServer(mockLogger, mockSearch, mockCfg)
			require.NoError(t, err)
			require.NotNil(t, server)

			// Create test request
			req := api.SearchRequest{
				Query: tt.query,
				Index: tt.index,
				Size:  tt.size,
			}
			body, err := json.Marshal(req)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			// Serve the request
			server.Handler.ServeHTTP(recorder, request)

			// Verify response
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response api.SearchResponse
			err = json.NewDecoder(recorder.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, tt.mockResults, response.Results)
			assert.Equal(t, int(tt.mockCount), response.Total)

			// Verify mock expectations
			mockLogger.AssertExpectations(t)
			mockSearch.AssertExpectations(t)
		})
	}
}
