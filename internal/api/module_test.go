// Package api_test provides tests for the API package.
package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	apitestutils "github.com/jonesrussell/gocrawl/internal/api/testutils"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	apitypes "github.com/jonesrussell/gocrawl/internal/types"
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
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := setupMockLogger()
	mockSearch := apitestutils.NewMockSearchManager()
	mockStorage := apitestutils.NewMockStorage()
	mockIndexManager := testutils.NewMockIndexManager()

	// Set up mock search expectations
	expectedQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "test",
			},
		},
		"size": 10,
	}
	mockSearch.On("Search", mock.Anything, "test", expectedQuery).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil)
	mockSearch.On("Count", mock.Anything, "test", expectedQuery).Return(int64(1), nil)
	mockSearch.On("Aggregate", mock.Anything, "test", mock.Anything).Return(map[string]any{}, nil)
	mockSearch.On("Close").Return(nil)

	// Create server config with security settings
	serverConfig := &config.ServerConfig{
		Address:      ":0", // Use random port for testing
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	serverConfig.Security.Enabled = true
	serverConfig.Security.APIKey = testAPIKey

	// Create mock config with test settings
	mockConfig := configtest.NewMockConfig().
		WithServerConfig(serverConfig).
		WithAppConfig(&config.AppConfig{
			Environment: "test",
		}).
		WithLogConfig(&config.LogConfig{
			Level: "info",
			Debug: false,
		}).
		WithElasticsearchConfig(&config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test_index",
		})

	// Set up mock storage expectations
	mockStorage.On("Search", mock.Anything, "test", mock.Anything).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil)
	mockStorage.On("Count", mock.Anything, "test", mock.Anything).Return(int64(1), nil)
	mockStorage.On("Close").Return(nil)

	// Set up mock index manager expectations
	mockIndexManager.On("EnsureIndex", mock.Anything, mock.Anything).Return(nil)
	mockIndexManager.On("DeleteIndex", mock.Anything, mock.Anything).Return(nil)
	mockIndexManager.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	mockIndexManager.On("UpdateMapping", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Store references for test assertions
	ts.logger = mockLogger

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
		TestAPIModule,
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
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	t.Run("returns ok status", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, healthEndpoint), http.NoBody)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
	})
}

// TestSearchEndpoint tests the search endpoint functionality.
func TestSearchEndpoint(t *testing.T) {
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	tests := []struct {
		name           string
		requestBody    string
		apiKey         string
		expectedStatus int
		expectedError  *apitypes.APIError
	}{
		{
			name:           "requires API key",
			requestBody:    `{"query": "test"}`,
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedError: &apitypes.APIError{
				Code:    http.StatusUnauthorized,
				Message: "API key is required",
			},
		},
		{
			name:           "returns search results with valid API key",
			requestBody:    `{"query": "test", "index": "test", "size": 10}`,
			apiKey:         testAPIKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "handles invalid JSON",
			requestBody:    `{"query": "test", invalid json}`,
			apiKey:         testAPIKey,
			expectedStatus: http.StatusBadRequest,
			expectedError: &apitypes.APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request payload",
			},
		},
		{
			name:           "handles empty query",
			requestBody:    `{"query": "", "index": "test"}`,
			apiKey:         testAPIKey,
			expectedStatus: http.StatusBadRequest,
			expectedError: &apitypes.APIError{
				Code:    http.StatusBadRequest,
				Message: "Query cannot be empty",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, searchEndpoint, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}

			w := httptest.NewRecorder()
			ts.server.Handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != nil {
				var errResp apitypes.APIError
				err := json.NewDecoder(w.Body).Decode(&errResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError.Code, errResp.Code)
				assert.Equal(t, tt.expectedError.Message, errResp.Message)
			} else if tt.expectedStatus == http.StatusOK {
				var resp apitypes.SearchResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Results)
				assert.Equal(t, 1, resp.Total)
			}
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
	mockConfig := configtest.NewMockConfig()
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
