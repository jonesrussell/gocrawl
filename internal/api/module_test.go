// Package api_test implements tests for the API package.
package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
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
	app    *fxtest.App
	server *http.Server
}

// setupMockSearchManager creates and configures a mock search manager for testing.
func setupMockSearchManager() *testutils.MockSearchManager {
	mockSearch := &testutils.MockSearchManager{}
	mockSearch.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]any{}, nil)
	mockSearch.On("Count", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)
	mockSearch.On("Aggregate", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockSearch.On("Close").Return(nil)
	return mockSearch
}

// setupMockLogger creates and configures a mock logger for testing.
func setupMockLogger() *testutils.MockLogger {
	mockLog := &testutils.MockLogger{}
	mockLog.On("Info", "StartHTTPServer function called", mock.Anything).Return()
	mockLog.On("Info", "Server started", mock.Anything).Return()
	mockLog.On("Info", "Server stopped", mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()
	return mockLog
}

func setupTest(t *testing.T) *testServer {
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := setupMockLogger()
	mockSearchManager := setupMockSearchManager()
	mockConfig := testutils.NewMockConfig()
	mockSecurityMiddleware := &testutils.MockSecurityMiddleware{}

	// Get the server instance
	var server *http.Server
	app := fxtest.New(t,
		fx.Supply(
			mockLogger,
			mockConfig,
			mockSecurityMiddleware,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			fx.Annotate(
				func() api.SearchManager { return mockSearchManager },
				fx.As(new(api.SearchManager)),
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

func (ts *testServer) cleanup() {
	if ts.app != nil {
		ts.app.RequireStop()
	}
}

func TestAPIModuleInitialization(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	assert.NotNil(t, ts.app, "Application should be initialized")
	assert.NotNil(t, ts.server, "HTTP server should be initialized")
}

func TestHealthEndpoint(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	t.Run("returns ok status", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, healthEndpoint), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
	})

	t.Run("sets security headers", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, healthEndpoint), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)

		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	})
}

func TestSearchEndpoint(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

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

	t.Run("enforces rate limiting", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), nil)
		require.NoError(t, err)
		req.Header.Set("X-Api-Key", testAPIKey)

		// Send valid requests up to rate limit
		for i := range 149 {
			w := httptest.NewRecorder()
			ts.server.Handler.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code, "Expected OK for request %d", i+1)
		}

		// Verify rate limit is triggered
		w := httptest.NewRecorder()
		ts.server.Handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "too many requests")
	})
}

func TestCORSConfiguration(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	req, err := http.NewRequest(http.MethodOptions, fmt.Sprintf("http://%s%s", ts.server.Addr, searchEndpoint), nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	w := httptest.NewRecorder()
	ts.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
}
