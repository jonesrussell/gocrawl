// Package api_test implements tests for the API package.
package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
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
	config        config.Interface
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

func setupTest(t *testing.T) *testServer {
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := setupMockLogger()
	mockConfig := testutils.NewMockConfig()
	mockSearch := setupMockSearchManager()

	// Store references for test assertions
	ts.logger = mockLogger
	ts.searchManager = mockSearch
	ts.config = mockConfig

	// Get the server instance
	var server *http.Server
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Supply(mockConfig),
		fx.Supply(mockLogger),
		fx.Supply(t.Context()),
		fx.Provide(
			func() api.SearchManager { return mockSearch },
		),
		fx.Replace(
			fx.Annotate(
				func(
					log common.Logger,
					searchManager api.SearchManager,
					cfg common.Config,
				) (*http.Server, middleware.SecurityMiddlewareInterface) {
					// Create router and security middleware
					router, security := api.SetupRouter(log, searchManager, cfg)

					// Create server
					httpServer := &http.Server{
						Addr:              cfg.GetServerConfig().Address,
						Handler:           router,
						ReadHeaderTimeout: api.ReadHeaderTimeout,
					}

					return httpServer, security
				},
				fx.As(new(*http.Server), new(middleware.SecurityMiddlewareInterface)),
			),
		),
		api.Module,
		fx.Invoke(func(s *http.Server) {
			server = s
		}),
	)

	ts.app = app
	app.RequireStart()
	ts.server = server

	return ts
}

func TestAPIModuleInitialization(t *testing.T) {
	ts := setupTest(t)
	defer ts.app.RequireStop()

	assert.NotNil(t, ts.app, "Application should be initialized")
	assert.NotNil(t, ts.server, "HTTP server should be initialized")
	assert.NotNil(t, ts.logger, "Logger should be initialized")
	assert.NotNil(t, ts.searchManager, "Search manager should be initialized")
	assert.NotNil(t, ts.config, "Config should be initialized")
}

func TestHealthEndpoint(t *testing.T) {
	ts := setupTest(t)
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

func TestSearchEndpoint(t *testing.T) {
	ts := setupTest(t)
	defer ts.app.RequireStop()

	t.Run("requires API key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "unauthorized")
	})

	t.Run("returns search results with valid API key", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), nil)
		require.NoError(t, err)
		req.Header.Set("X-Api-Key", testAPIKey)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "results")
	})
}
