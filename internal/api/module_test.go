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
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/testutils"
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
	logger common.Logger
}

// setupMockLogger creates and configures a mock logger for testing.
func setupMockLogger() *testutils.MockLogger {
	mockLog := testutils.NewMockLogger()
	mockLog.On("Info", mock.Anything, mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()
	mockLog.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLog.On("Printf", mock.Anything, mock.Anything).Return()
	mockLog.On("Errorf", mock.Anything, mock.Anything).Return()
	mockLog.On("Sync").Return(nil)
	return mockLog
}

// TestAPIModule provides a test version of the API module with test-specific dependencies.
var TestAPIModule = fx.Module("testAPI",
	// Suppress fx logging to reduce noise in the application logs.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),
	// Core modules used by most commands, excluding logger and sources.
	config.TestModule,
	// Include the API module itself
	api.Module,
)

func setupTestApp(t *testing.T) *testServer {
	ts := &testServer{}

	// Set environment variables for test configuration
	t.Setenv("CRAWLER_SOURCE_FILE", "testdata/sources.yml")
	t.Setenv("CONFIG_FILE", "testdata/config.yaml")

	// Create mock dependencies
	mockLogger := setupMockLogger()
	mockSearch := testutils.NewMockSearchManager()
	mockConfig := configtest.NewMockConfig().
		WithServerConfig(&config.ServerConfig{
			Address:      ":0", // Use random port for testing
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
	mockSearch.On("Close").Return(nil)

	// Store references for test assertions
	ts.logger = mockLogger

	// Create and start the application
	app := fxtest.New(t,
		TestAPIModule,
		fx.NopLogger,
		fx.Supply(
			mockConfig,
			fx.Annotate(
				mockSearch,
				fx.As(new(api.SearchManager)),
			),
			fx.Annotate(
				t.Context(),
				fx.As(new(context.Context)),
			),
		),
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
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, healthEndpoint), nil)
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
				log types.Logger,
				searchManager api.SearchManager,
				cfg common.Config,
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
				fx.As(new(common.Config)),
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
				fx.As(new(types.Logger)),
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
