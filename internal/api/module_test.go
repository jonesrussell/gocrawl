// Package api_test provides tests for the API package.
package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
	app           *fxtest.App
	server        *http.Server
	logger        common.Logger
	searchManager api.SearchManager
	config        common.Config
}

// setupMockSearchManager creates and configures a mock search manager for testing.
func setupMockSearchManager() *testutils.MockSearchManager {
	mockSearch := &testutils.MockSearchManager{}
	mockSearch.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]any{
		map[string]any{
			"title":   "Test Result",
			"url":     "https://test.com",
			"content": "This is a test result",
		},
	}, nil)
	mockSearch.On("Count", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockSearch.On("Aggregate", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockSearch.On("Close").Return(nil)
	return mockSearch
}

// setupMockLogger creates and configures a mock logger for testing.
func setupMockLogger() *testutils.MockLogger {
	mockLog := &testutils.MockLogger{}
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

// TestCommonModule provides a test-specific common module that excludes the logger module.
var TestCommonModule = fx.Module("testCommon",
	// Suppress fx logging to reduce noise in the application logs.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),
	// Core modules used by most commands, excluding config and sources.
	logger.Module,
)

// TestAPIModule provides a test version of the API module with test-specific dependencies.
var TestAPIModule = fx.Module("api",
	TestCommonModule,
	fx.Provide(
		// Provide test context
		context.Background,
		// Provide mock search manager
		func() api.SearchManager {
			return setupMockSearchManager()
		},
		// Provide mock config
		fx.Annotate(
			func() common.Config {
				mockCfg := &testutils.MockConfig{}
				mockCfg.On("GetAppConfig").Return(&config.AppConfig{
					Environment: "test",
					Name:        "gocrawl",
					Version:     "1.0.0",
					Debug:       true,
				})
				mockCfg.On("GetLogConfig").Return(&config.LogConfig{
					Level: "debug",
					Debug: true,
				})
				mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
					IndexName: "test-index",
				})
				mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				return mockCfg
			},
			fx.As(new(common.Config)),
		),
		// Provide the server and security middleware
		func(
			log common.Logger,
			searchManager api.SearchManager,
			cfg common.Config,
		) (*http.Server, middleware.SecurityMiddlewareInterface) {
			// Create router and security middleware
			router, security := api.SetupRouter(log, searchManager, cfg)

			// Create server
			server := &http.Server{
				Addr:              cfg.GetServerConfig().Address,
				Handler:           router,
				ReadHeaderTimeout: api.ReadHeaderTimeout,
			}

			return server, security
		},
	),
	fx.Invoke(api.ConfigureLifecycle),
)

func setupTestApp(t *testing.T) *testServer {
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := setupMockLogger()

	// Store references for test assertions
	ts.logger = mockLogger

	// Create and start the application
	app := fxtest.New(t,
		TestAPIModule,
		fx.NopLogger,
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

	t.Run("requires API key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), strings.NewReader(`{"query":"test","index":"test","size":10}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"Unauthorized"}`, w.Body.String())
	})

	t.Run("returns search results with valid API key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), strings.NewReader(`{"query":"test","index":"test","size":10}`))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", testAPIKey)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "results")
	})
}
