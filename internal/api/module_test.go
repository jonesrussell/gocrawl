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

func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	// Create server config that will be used multiple times
	serverCfg := &config.ServerConfig{
		Address:      ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Set up mocks with required expectations
	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
	mockLogger.EXPECT().Info("Server configuration", "address", ":8080").Times(1)
	mockLogger.EXPECT().Info("HTTP server started", "address", ":8080").Times(1)
	mockLogger.EXPECT().Info("Shutting down HTTP server...").Times(1)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes() // For potential errors

	mockCfg := NewMockConfig(ctrl)
	mockCfg.EXPECT().GetServerConfig().Return(serverCfg).Times(2)

	mockSearch := NewMockSearchManager(ctrl)

	var server *http.Server
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.SearchManager { return mockSearch },
			func() config.Interface { return mockCfg },
		),
		api.Module,
		fx.Invoke(func(s *http.Server) {
			server = s
		}),
	)

	require.NoError(t, app.Err())
	app.RequireStart()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		if server != nil {
			server.Shutdown(ctx)
		}
		app.RequireStop()
	})

	require.NotNil(t, server, "HTTP server should be initialized")
	assert.Equal(t, ":8080", server.Addr, "Server should listen on port 8080")
	assert.Equal(t, serverCfg.ReadTimeout, server.ReadTimeout, "Read timeout should match config")
	assert.Equal(t, serverCfg.WriteTimeout, server.WriteTimeout, "Write timeout should match config")
	assert.Equal(t, serverCfg.IdleTimeout, server.IdleTimeout, "Idle timeout should match config")
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
			t.Cleanup(func() { ctrl.Finish() })

			// Create server config
			serverCfg := &config.ServerConfig{
				Address:      ":8080",
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			// Set up mocks
			mockLogger := logger.NewMockInterface(ctrl)
			mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
			mockLogger.EXPECT().Info("Server configuration", "address", ":8080").Times(1)
			mockLogger.EXPECT().Info("HTTP server started", "address", ":8080").Times(1)
			mockLogger.EXPECT().Info("Shutting down HTTP server...").Times(1)
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			mockCfg := NewMockConfig(ctrl)
			mockCfg.EXPECT().GetServerConfig().Return(serverCfg).Times(2)

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
				t.Cleanup(tt.cleanup)
			}

			// Set environment variable if specified
			if tt.envPort != "" {
				t.Setenv("GOCRAWL_PORT", tt.envPort)
			}

			ctrl := gomock.NewController(t)
			t.Cleanup(func() { ctrl.Finish() })

			// Create server config
			serverCfg := &config.ServerConfig{
				Address:      tt.configPort,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			// Set up mocks
			mockLogger := logger.NewMockInterface(ctrl)
			mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
			mockLogger.EXPECT().Info("Server configuration", "address", tt.expectedAddr).Times(1)
			mockLogger.EXPECT().Info("HTTP server started", "address", tt.expectedAddr).Times(1)
			mockLogger.EXPECT().Info("Shutting down HTTP server...").Times(1)
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			mockCfg := NewMockConfig(ctrl)
			mockCfg.EXPECT().GetServerConfig().Return(serverCfg).Times(2)

			mockSearch := NewMockSearchManager(ctrl)

			// Start the server
			server, err := api.StartHTTPServer(mockLogger, mockSearch, mockCfg)
			require.NoError(t, err)
			require.NotNil(t, server)

			// Verify the server address
			assert.Equal(t, tt.expectedAddr, server.Addr)
		})
	}
}
