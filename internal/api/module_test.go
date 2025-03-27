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

type testServer struct {
	app    *fxtest.App
	server *http.Server
}

func setupTest(t *testing.T) *testServer {
	ts := &testServer{}

	// Create mock dependencies
	mockLogger := &testutils.MockLogger{}
	mockSearchManager := &testutils.MockSearchManager{}
	mockConfig := testutils.NewMockConfig()
	mockSecurityMiddleware := &testutils.MockSecurityMiddleware{}

	// Set up mock expectations
	mockLogger.On("Info", "StartHTTPServer function called", mock.Anything).Return()
	mockLogger.On("Info", "Server started", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockSearchManager.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]any{}, nil)
	mockSearchManager.On("Count", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)
	mockSearchManager.On("Aggregate", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mockSearchManager.On("Close").Return(nil)

	// Get the server instance
	var server *http.Server
	app := fxtest.New(t,
		fx.Supply(
			mockLogger,
			mockConfig,
			mockSecurityMiddleware,
		),
		fx.Provide(
			// Provide context
			func() context.Context { return t.Context() },
			// Provide search manager with correct type
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

	// Start the app
	app.RequireStart()
	ts.server = server

	return ts
}

func (ts *testServer) cleanup() {
	if ts.app != nil {
		ts.app.RequireStop()
	}
}

func TestModuleConstruction(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Verify the app started successfully
	assert.NotNil(t, ts.app)
	assert.NotNil(t, ts.server)
}

func TestHealthHandler(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/health", ts.server.Addr), nil)
	require.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request
	ts.server.Handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestSearchHandler(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/search", ts.server.Addr), nil)
	require.NoError(t, err)

	// Add API key header
	req.Header.Set("X-Api-Key", "test-key")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request
	ts.server.Handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "results")
}

func TestAPIKeyAuthentication(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request without API key
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/search", ts.server.Addr), nil)
	require.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request
	ts.server.Handler.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestRateLimiting(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request with API key
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/search", ts.server.Addr), nil)
	require.NoError(t, err)
	req.Header.Set("X-Api-Key", "test-key")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send multiple requests
	for i := 0; i < 150; i++ {
		ts.server.Handler.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			break
		}
		w = httptest.NewRecorder()
	}

	// Verify rate limiting
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "too many requests")
}

func TestCORS(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request with CORS headers
	req, err := http.NewRequest(http.MethodOptions, fmt.Sprintf("http://%s/search", ts.server.Addr), nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request
	ts.server.Handler.ServeHTTP(w, req)

	// Verify CORS headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
}

func TestSecurityHeaders(t *testing.T) {
	ts := setupTest(t)
	defer ts.cleanup()

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/health", ts.server.Addr), nil)
	require.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request
	ts.server.Handler.ServeHTTP(w, req)

	// Verify security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}
