// Package api_test provides tests for the API package.
package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	apimocks "github.com/jonesrussell/gocrawl/testutils/mocks/api"
	configmocks "github.com/jonesrussell/gocrawl/testutils/mocks/config"
	loggermocks "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	storagemocks "github.com/jonesrussell/gocrawl/testutils/mocks/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
)

const (
	testAPIKey     = "test-key"
	healthEndpoint = "/health"
	searchEndpoint = "/search"
)

// testServer wraps the test application and server for API tests.
type testServer struct {
	app    *fxtest.App
	server *http.Server
	logger logger.Interface
}

// setupMockLogger creates and configures a mock logger for testing.
func setupMockLogger(ctrl *gomock.Controller) logger.Interface {
	mockLog := loggermocks.NewMockInterface(ctrl)
	mockLog.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Fatal(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().With(gomock.Any()).Return(mockLog).AnyTimes()
	return mockLog
}

// TestAPIModule provides a test version of the API module with test-specific dependencies.
var TestAPIModule = fx.Module("testAPI",
	// Suppress fx logging to reduce noise in the application logs.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),
	// Include only the necessary components from the API module
	fx.Provide(
		// Provide the server and security middleware together to avoid circular dependencies
		func(
			cfg config.Interface,
			log logger.Interface,
			searchManager api.SearchManager,
		) (*http.Server, middleware.SecurityMiddlewareInterface, error) {
			return api.StartHTTPServer(log, searchManager, cfg)
		},
		api.NewLifecycle,
		api.NewServer,
	),
)

func setupTestApp(t *testing.T) *testServer {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	// Create mock dependencies
	mockConfig := configmocks.NewMockInterface(ctrl)
	mockConfig.EXPECT().GetAppConfig().Return(&app.Config{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	}).AnyTimes()
	mockConfig.EXPECT().GetLogConfig().Return(&log.Config{
		Level:      "debug",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}).AnyTimes()
	mockConfig.EXPECT().GetElasticsearchConfig().Return(&elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	}).AnyTimes()
	mockConfig.EXPECT().GetServerConfig().Return(&server.Config{
		SecurityEnabled: true,
		APIKey:          testAPIKey,
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}).AnyTimes()
	mockConfig.EXPECT().GetSources().Return([]types.Source{}).AnyTimes()
	mockConfig.EXPECT().GetCommand().Return("test").AnyTimes()
	mockConfig.EXPECT().GetPriorityConfig().Return(&priority.Config{
		DefaultPriority: 1,
		Rules:           []priority.Rule{},
	}).AnyTimes()

	mockLogger := setupMockLogger(ctrl)
	mockSearch := apimocks.NewMockSearchManager(ctrl)
	mockStorage := storagemocks.NewMockInterface(ctrl)
	mockIndexManager := apimocks.NewMockIndexManager(ctrl)

	// Set up mock search expectations
	expectedQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "test",
			},
		},
		"size": 10,
	}
	t.Logf("Setting up mock expectations with query: %+v", expectedQuery)
	mockSearch.EXPECT().Search(gomock.Any(), "", expectedQuery).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil).AnyTimes()

	countQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "test",
			},
		},
		"size": 10,
	}
	mockSearch.EXPECT().Count(gomock.Any(), "", countQuery).Return(int64(1), nil).AnyTimes()

	ts := &testServer{
		logger: mockLogger,
	}

	// Create and start the application
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Supply(
			fx.Annotate(
				mockConfig,
				fx.As(new(config.Interface)),
			),
			fx.Annotate(
				mockSearch,
				fx.As(new(api.SearchManager)),
			),
			fx.Annotate(
				mockLogger,
				fx.As(new(logger.Interface)),
			),
			fx.Annotate(
				mockStorage,
				fx.As(new(storagetypes.Interface)),
			),
			fx.Annotate(
				mockIndexManager,
				fx.As(new(api.IndexManager)),
			),
			fx.Annotate(
				t.Context(),
				fx.As(new(context.Context)),
			),
		),
		api.Module,
		fx.Invoke(func(s *http.Server) {
			ts.server = s
		}),
	)

	ts.app = app
	app.RequireStart()

	return ts
}

// TestAPIModuleInitialization verifies that the API module can be initialized with all required dependencies.
func TestAPIModuleInitialization(t *testing.T) {
	t.Parallel()
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	assert.NotNil(t, ts.app, "Application should be initialized")
	assert.NotNil(t, ts.server, "HTTP server should be initialized")
}

// TestHealthEndpoint verifies that the health endpoint works correctly.
func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	// Create test server
	ts := setupTestApp(t)

	// Set up mock expectations for logger
	mockLogger := ts.logger.(*loggermocks.MockInterface)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Return()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "returns ok status",
			method:         "GET",
			path:           "/health",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			// Send request
			ts.server.Handler.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestSearchEndpoint tests the search endpoint functionality.
func TestSearchEndpoint(t *testing.T) {
	t.Parallel()

	// Create test server
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	// Set up mock expectations for logger
	mockLogger := ts.logger.(*loggermocks.MockInterface)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Return().AnyTimes()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		apiKey         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "requires API key",
			method:         "POST",
			path:           "/search",
			body:           `{"query": "test"}`,
			apiKey:         "",
			expectedStatus: 401,
			expectedBody:   `{"code":401,"message":"missing API key"}`,
		},
		{
			name:           "handles invalid API key",
			method:         "POST",
			path:           "/search",
			body:           `{"query": "test"}`,
			apiKey:         "wrong-key",
			expectedStatus: 401,
			expectedBody:   `{"code":401,"message":"invalid API key"}`,
		},
		{
			name:           "handles invalid JSON",
			method:         "POST",
			path:           "/search",
			body:           "invalid json",
			apiKey:         testAPIKey,
			expectedStatus: 400,
			expectedBody:   `{"code":400,"message":"Invalid request payload"}`,
		},
		{
			name:           "handles empty query",
			method:         "POST",
			path:           "/search",
			body:           `{"query": ""}`,
			apiKey:         testAPIKey,
			expectedStatus: 400,
			expectedBody:   `{"code":400,"message":"Query cannot be empty"}`,
		},
		{
			name:           "returns search results with valid request",
			method:         "POST",
			path:           "/search",
			body:           `{"query": "test"}`,
			apiKey:         testAPIKey,
			expectedStatus: 200,
			expectedBody:   `{"results":[{"title":"Test Result","url":"https://test.com"}],"total":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create request with body
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()

			// Send request
			ts.server.Handler.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

// TestLoggerDependencyRegression verifies that the logger dependency is properly injected.
func TestLoggerDependencyRegression(t *testing.T) {
	t.Parallel()
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	// Verify that the logger is properly injected
	assert.NotNil(t, ts.logger, "Logger should be properly injected")
}

// TestModule tests the API module.
func TestModule(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	// Create mock dependencies
	mockConfig := configmocks.NewMockInterface(ctrl)
	mockLogger := setupMockLogger(ctrl)
	mockStorage := storagemocks.NewMockInterface(ctrl)
	mockIndexManager := apimocks.NewMockIndexManager(ctrl)

	// Set up mock storage expectations
	mockStorage.EXPECT().GetIndexDocCount(gomock.Any(), gomock.Any()).Return(int64(0), nil).AnyTimes()
	mockStorage.EXPECT().TestConnection(gomock.Any()).Return(nil).AnyTimes()
	mockStorage.EXPECT().Close().Return(nil).AnyTimes()

	// Set up mock index manager expectations
	mockIndexManager.EXPECT().EnsureIndex(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockIndexManager.EXPECT().DeleteIndex(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockIndexManager.EXPECT().IndexExists(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	mockIndexManager.EXPECT().GetMapping(gomock.Any(), gomock.Any()).Return(map[string]any{}, nil).AnyTimes()
	mockIndexManager.EXPECT().UpdateMapping(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Set up mock config expectations
	mockConfig.EXPECT().GetAppConfig().Return(&app.Config{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	}).AnyTimes()
	mockConfig.EXPECT().GetLogConfig().Return(&log.Config{
		Level:      "debug",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}).AnyTimes()
	mockConfig.EXPECT().GetElasticsearchConfig().Return(&elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	}).AnyTimes()
	mockConfig.EXPECT().GetServerConfig().Return(&server.Config{
		SecurityEnabled: true,
		APIKey:          "test-key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}).AnyTimes()
	mockConfig.EXPECT().GetSources().Return([]types.Source{}).AnyTimes()
	mockConfig.EXPECT().GetCommand().Return("test").AnyTimes()
	mockConfig.EXPECT().GetPriorityConfig().Return(&priority.Config{
		DefaultPriority: 1,
		Rules:           []priority.Rule{},
	}).AnyTimes()

	app := fx.New(
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface { return mockConfig },
			func() logger.Interface { return mockLogger },
			func() storagetypes.Interface { return mockStorage },
			func() api.IndexManager { return mockIndexManager },
		),
		api.Module,
	)

	err := app.Start(t.Context())
	require.NoError(t, err)

	err = app.Stop(t.Context())
	require.NoError(t, err)
}
