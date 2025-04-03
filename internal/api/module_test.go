// Package api_test provides tests for the API package.
package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/api/testutils"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
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
func setupMockLogger() *testutils.MockLogger {
	mockLog := testutils.NewMockLogger()
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
		api.NewServer,
		api.NewLifecycle,
	),
)

func setupTestApp(t *testing.T) *testServer {
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := setupMockLogger()
	mockSearch := testutils.NewMockSearchManager()
	mockStorage := testutils.NewMockStorage()
	mockIndexManager := testutils.NewMockIndexManager()

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

	// Set up mock search expectations
	mockSearch.On("Search", mock.Anything, "test", mock.Anything).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil)
	mockSearch.On("Count", mock.Anything, "test", mock.Anything).Return(int64(1), nil)
	mockSearch.On("Aggregate", mock.Anything, "test", mock.Anything).Return(map[string]any{}, nil)
	mockSearch.On("Close").Return(nil)

	// Set up mock storage expectations
	mockStorage.On("Search", mock.Anything, "test", mock.Anything, mock.Anything).Return([]any{
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
	mockIndexManager.On("GetMapping", mock.Anything, mock.Anything).Return(map[string]any{}, nil)
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
				fx.As(new(interfaces.IndexManager)),
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

// TestSearchEndpoint verifies that the search endpoint works correctly.
func TestSearchEndpoint(t *testing.T) {
	ts := setupTestApp(t)
	defer ts.app.RequireStop()

	reqBody := `{"query":"test","index":"test","size":10}`
	endpoint := fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint)

	t.Run("requires API key", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			endpoint,
			strings.NewReader(reqBody),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
	})

	t.Run("returns search results with valid API key", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			endpoint,
			strings.NewReader(reqBody),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", testAPIKey)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "results")
		assert.Contains(t, w.Body.String(), "Test Result")
		assert.Contains(t, w.Body.String(), "https://test.com")
	})
}

// TestLoggerDependencyRegression verifies that the logger dependency is properly provided and used.
func TestLoggerDependencyRegression(t *testing.T) {
	// Create a mock logger that we can verify is used
	mockLog := setupMockLogger()

	// Create test configuration
	mockConfig := configtest.NewMockConfig().
		WithServerConfig(&config.ServerConfig{
			Address:      ":0",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}).
		WithAppConfig(&config.AppConfig{
			Environment: "test",
		}).
		WithLogConfig(&config.LogConfig{
			Level: "info",
			Debug: false,
		})

	// Create mock search manager
	mockSearch := testutils.NewMockSearchManager()
	mockSearch.On("Search", mock.Anything, "test", mock.Anything).Return([]any{
		map[string]any{
			"title": "Test Result",
			"url":   "https://test.com",
		},
	}, nil)
	mockSearch.On("Count", mock.Anything, "test", mock.Anything).Return(int64(1), nil)
	mockSearch.On("Close").Return(nil)

	// Create a test module that only includes the necessary components
	testModule := fx.Module("testAPI",
		fx.Provide(
			// Provide the server and security middleware together to avoid circular dependencies
			func(
				log logger.Interface,
				searchManager api.SearchManager,
				cfg config.Interface,
			) (*http.Server, middleware.SecurityMiddlewareInterface) {
				// Use StartHTTPServer to create the server and security middleware
				server, security, err := api.StartHTTPServer(log, searchManager, cfg)
				if err != nil {
					panic(err)
				}
				return server, security
			},
		),
		fx.Invoke(api.ConfigureLifecycle),
	)

	// Create test application with just the test module
	app := fxtest.New(t,
		testModule,
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
				t.Context(),
				fx.As(new(context.Context)),
			),
			fx.Annotate(
				mockLog,
				fx.As(new(logger.Interface)),
			),
		),
	)

	// Start the application
	app.RequireStart()
	defer app.RequireStop()

	// Verify that the logger was used by checking if the expected log calls were made
	mockLog.AssertCalled(t, "Info", "StartHTTPServer function called", mock.Anything)
	mockLog.AssertCalled(t, "Info", "Server configuration", mock.Anything)
}

// TestModule tests the API module.
func TestModule(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockConfig := configtest.NewMockConfig()
	mockLogger := testutils.NewMockLogger()
	mockStorage := testutils.NewMockStorage()
	mockIndexManager := testutils.NewMockIndexManager()

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
			func() interfaces.IndexManager { return mockIndexManager },
		),
		api.Module,
	)

	err := app.Start(t.Context())
	require.NoError(t, err)

	err = app.Stop(t.Context())
	require.NoError(t, err)
}
