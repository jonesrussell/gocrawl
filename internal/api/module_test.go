package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const (
	defaultPort = ":8080" // Default port matching the one in module.go
)

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	ctrl     *gomock.Controller
	recorder *mockSearchManagerMockRecorder
}

type mockSearchManagerMockRecorder struct {
	mock *mockSearchManager
}

func NewMockSearchManager(ctrl *gomock.Controller) *mockSearchManager {
	mock := &mockSearchManager{ctrl: ctrl}
	mock.recorder = &mockSearchManagerMockRecorder{mock}
	return mock
}

func (m *mockSearchManager) EXPECT() *mockSearchManagerMockRecorder {
	return m.recorder
}

func (m *mockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", ctx, index, query)
	ret0, _ := ret[0].([]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (m *mockSearchManagerMockRecorder) Search(ctx, index, query interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Search", nil, ctx, index, query)
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Count", ctx, index, query)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (m *mockSearchManagerMockRecorder) Count(ctx, index, query interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Count", nil, ctx, index, query)
}

func (m *mockSearchManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Aggregate", ctx, index, aggs)
	ret1, _ := ret[1].(error)
	return ret[0], ret1
}

func (m *mockSearchManagerMockRecorder) Aggregate(ctx, index, aggs interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Aggregate", nil, ctx, index, aggs)
}

// mockConfig implements config.Interface for testing
type mockConfig struct {
	ctrl     *gomock.Controller
	recorder *mockConfigMockRecorder
}

type mockConfigMockRecorder struct {
	mock *mockConfig
}

func NewMockConfig(ctrl *gomock.Controller) *mockConfig {
	mock := &mockConfig{ctrl: ctrl}
	mock.recorder = &mockConfigMockRecorder{mock}
	return mock
}

func (m *mockConfig) EXPECT() *mockConfigMockRecorder {
	return m.recorder
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerConfig")
	ret0, _ := ret[0].(*config.ServerConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetServerConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetServerConfig", nil)
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetElasticsearchConfig")
	ret0, _ := ret[0].(*config.ElasticsearchConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetElasticsearchConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetElasticsearchConfig", nil)
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCrawlerConfig")
	ret0, _ := ret[0].(*config.CrawlerConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetCrawlerConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetCrawlerConfig", nil)
}

func (m *mockConfig) GetLogConfig() *config.LogConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogConfig")
	ret0, _ := ret[0].(*config.LogConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetLogConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetLogConfig", nil)
}

func (m *mockConfig) GetAppConfig() *config.AppConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppConfig")
	ret0, _ := ret[0].(*config.AppConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetAppConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetAppConfig", nil)
}

func (m *mockConfig) GetSources() []config.Source {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSources")
	ret0, _ := ret[0].([]config.Source)
	return ret0
}

func (m *mockConfigMockRecorder) GetSources() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetSources", nil)
}

func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := NewMockConfig(ctrl)

	// Set up expectations for required config methods
	mockCfg.EXPECT().GetElasticsearchConfig().Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "changeme",
	}).AnyTimes()

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
		),
		api.Module,
	)
	require.NoError(t, app.Err())
}

func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := NewMockConfig(ctrl)

	// Set up expectations for required config methods
	mockCfg.EXPECT().GetElasticsearchConfig().Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "changeme",
	}).AnyTimes()

	var searchManager api.SearchManager

	app := fxtest.New(t,
		api.Module,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
		),
		fx.Populate(&searchManager),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, searchManager)
}

func TestStartHTTPServer(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		index          string
		size           int
		mockResults    []any
		mockCount      int64
		expectedStatus int
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "test query",
			index: "articles",
			size:  10,
			mockResults: []any{
				map[string]any{
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
			mockResults:    []any{},
			mockCount:      0,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock logger
			mockLogger := logger.NewMockInterface(ctrl)
			mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
			mockLogger.EXPECT().Info("Server configuration", "address", ":8080").Times(1)

			// Create mock search manager
			mockSearch := NewMockSearchManager(ctrl)

			// Set up expected queries
			searchQuery := map[string]any{
				"query": map[string]any{
					"match": map[string]any{
						"content": tt.query,
					},
				},
				"size": tt.size,
			}

			countQuery := map[string]any{
				"query": map[string]any{
					"match": map[string]any{
						"content": tt.query,
					},
				},
			}

			// Set up mock expectations
			mockSearch.EXPECT().Search(gomock.Any(), tt.index, searchQuery).Return(tt.mockResults, nil).Times(1)
			mockSearch.EXPECT().Count(gomock.Any(), tt.index, countQuery).Return(tt.mockCount, nil).Times(1)

			mockCfg := NewMockConfig(ctrl)
			mockCfg.EXPECT().GetServerConfig().Return(&config.ServerConfig{
				Address:      ":8080",
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}).Times(1)

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
		})
	}
}

func TestStartHTTPServer_PortConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		configPort   string
		envPort      string
		expectedAddr string
		cleanup      func()
	}{
		{
			name:         "use config port",
			configPort:   ":3000",
			expectedAddr: ":3000",
		},
		{
			name:         "use env port when config empty",
			configPort:   "",
			envPort:      "4000",
			expectedAddr: ":4000",
			cleanup: func() {
				os.Unsetenv("GOCRAWL_PORT")
			},
		},
		{
			name:         "use env port with colon",
			configPort:   "",
			envPort:      ":5000",
			expectedAddr: ":5000",
			cleanup: func() {
				os.Unsetenv("GOCRAWL_PORT")
			},
		},
		{
			name:         "use default port when no config or env",
			configPort:   "",
			expectedAddr: defaultPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// Set environment variable if specified
			if tt.envPort != "" {
				t.Setenv("GOCRAWL_PORT", tt.envPort)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock logger
			mockLogger := logger.NewMockInterface(ctrl)
			mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
			mockLogger.EXPECT().Info("Server configuration", "address", tt.expectedAddr).Times(1)

			// Create mock search manager
			mockSearch := NewMockSearchManager(ctrl)

			mockCfg := NewMockConfig(ctrl)
			mockCfg.EXPECT().GetServerConfig().Return(&config.ServerConfig{
				Address:      tt.configPort,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}).Times(1)

			// Start the server
			server, err := api.StartHTTPServer(mockLogger, mockSearch, mockCfg)
			require.NoError(t, err)
			require.NotNil(t, server)

			// Verify the server address
			assert.Equal(t, tt.expectedAddr, server.Addr)
		})
	}
}
