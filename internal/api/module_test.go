// Package api_test provides tests for the API package.
package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	apitestutils "github.com/jonesrussell/gocrawl/internal/api/testutils"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
func setupMockLogger() *apitestutils.MockLogger {
	mockLog := apitestutils.NewMockLogger()
	mockLog.On("Info", mock.Anything, mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()
	mockLog.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLog.On("With", mock.Anything).Return(mockLog)
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

	// Create mock dependencies
	mockConfig := &configtestutils.MockConfig{}
	mockConfig.On("GetAppConfig").Return(&config.AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockConfig.On("GetLogConfig").Return(&config.LogConfig{
		Level: "debug",
		Debug: true,
	})
	mockConfig.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	})
	mockConfig.On("GetServerConfig").Return(&config.ServerConfig{
		Address: ":8080",
		Security: struct {
			Enabled   bool   `yaml:"enabled"`
			APIKey    string `yaml:"api_key"`
			RateLimit int    `yaml:"rate_limit"`
			CORS      struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			} `yaml:"cors"`
			TLS config.TLSConfig `yaml:"tls"`
		}{
			Enabled:   true,
			APIKey:    testAPIKey,
			RateLimit: 100,
		},
	})
	mockConfig.On("GetSources").Return([]config.Source{}, nil)
	mockConfig.On("GetCommand").Return("test")
	mockConfig.On("GetPriorityConfig").Return(&config.PriorityConfig{
		Default: 1,
		Rules:   []config.PriorityRule{},
	})
	mockLogger := apitestutils.NewMockLogger()
	mockSearch := apitestutils.NewMockSearchManager()
	mockStorage := apitestutils.NewMockStorage()
	mockIndexManager := apitestutils.NewMockIndexManager()

	// Set up mock search expectations
	expectedQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "test",
			},
		},
		"size": 0,
	}
	t.Logf("Setting up mock expectations with query: %+v", expectedQuery)
	mockSearch.On("Search", t.Context(), "", expectedQuery).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil).Run(func(args mock.Arguments) {
		t.Logf("Search called with args: %+v", args)
	})
	mockSearch.On("Count", t.Context(), "", expectedQuery).Return(int64(1), nil).Run(func(args mock.Arguments) {
		t.Logf("Count called with args: %+v", args)
	})

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
				fx.As(new(types.Interface)),
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
	mockLogger := ts.logger.(*apitestutils.MockLogger)
	mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

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
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
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

	// Set up mock expectations for logger
	mockLogger := ts.logger.(*apitestutils.MockLogger)
	mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

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
			// Create request
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
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	// Verify that the logger is properly injected
	assert.NotNil(t, ts.logger, "Logger should be properly injected")
}

// TestModule tests the API module.
func TestModule(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockConfig := &configtestutils.MockConfig{}
	mockLogger := apitestutils.NewMockLogger()
	mockStorage := apitestutils.NewMockStorage()
	mockIndexManager := apitestutils.NewMockIndexManager()

	// Set up mock storage expectations
	mockStorage.On("GetIndexDocCount", mock.Anything, mock.Anything).Return(int64(0), nil)
	mockStorage.On("Ping", mock.Anything).Return(nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("Close").Return(nil)

	// Set up mock index manager expectations
	mockIndexManager.On("EnsureIndex", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockIndexManager.On("DeleteIndex", mock.Anything, mock.Anything).Return(nil)
	mockIndexManager.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	mockIndexManager.On("GetMapping", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
	mockIndexManager.On("UpdateMapping", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app := fx.New(
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface { return mockConfig },
			func() logger.Interface { return mockLogger },
			func() types.Interface { return mockStorage },
			func() api.IndexManager { return mockIndexManager },
		),
		api.Module,
	)

	err := app.Start(t.Context())
	require.NoError(t, err)

	err = app.Stop(t.Context())
	require.NoError(t, err)
}

func TestSecurityMiddleware(t *testing.T) {
	// Create test dependencies
	serverConfig := &config.ServerConfig{}
	serverConfig.Security.Enabled = true
	serverConfig.Security.APIKey = testAPIKey
	mockLogger := setupMockLogger()

	// Create security middleware
	securityMiddleware := middleware.NewSecurityMiddleware(serverConfig, mockLogger)

	// Create Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityMiddleware.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name     string
		apiKey   string
		expected int
	}{
		{"valid key", testAPIKey, http.StatusOK},
		{"invalid key", "wrong-key", http.StatusUnauthorized},
		{"missing key", "", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expected, w.Code)
		})
	}
}
