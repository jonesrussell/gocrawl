package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	loggerMock "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx/fxevent"
)

// mockLogger implements common.Logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *mockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	if ret, ok := args.Get(0).(logger.Interface); ok {
		return ret
	}
	return m
}

func (m *mockLogger) NewFxLogger() fxevent.Logger {
	args := m.Called()
	if ret, ok := args.Get(0).(fxevent.Logger); ok {
		return ret
	}
	return &fxevent.NopLogger
}

// mockTimeProvider is a mock implementation of TimeProvider
type mockTimeProvider struct {
	currentTime time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *mockTimeProvider) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func setupTestRouter(t *testing.T, cfg *server.Config) (*gin.Engine, *middleware.SecurityMiddleware, *metrics.Metrics, *mockTimeProvider) {
	t.Helper()

	// Create metrics
	m := metrics.NewMetrics()

	// Create security middleware
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })
	mockLog := loggerMock.NewMockInterface(ctrl)
	mockLog.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Fatal(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().With(gomock.Any()).Return(mockLog).AnyTimes()

	security := middleware.NewSecurityMiddleware(cfg, mockLog)
	security.SetMetrics(m)

	// Set up mock time provider
	mockTime := &mockTimeProvider{
		currentTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	security.SetTimeProvider(mockTime)
	security.SetRateLimitWindow(5 * time.Second)
	security.SetMaxRequests(2)

	// Create router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(security.Middleware())
	router.GET("/test", func(c *gin.Context) {
		t.Log("Handling test request")
		c.Status(http.StatusOK)
	})

	return router, security, m, mockTime
}

func TestAPIKeyAuthentication(t *testing.T) {
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, _, m, _ := setupTestRouter(t, cfg)

	// Test missing API key
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Request without API key should fail")

	// Test invalid API key
	req.Header.Set("X-Api-Key", "invalid:key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Request with invalid API key should fail")

	// Test valid API key
	req.Header.Set("X-Api-Key", "test:key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Request with valid API key should succeed")

	// Verify metrics
	assert.Equal(t, int64(1), m.GetSuccessfulRequests(), "Should have 1 successful request")
	assert.Equal(t, int64(2), m.GetFailedRequests(), "Should have 2 failed requests")
}

func TestRateLimiting(t *testing.T) {
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, security, m, mockTime := setupTestRouter(t, cfg)

	// Set a shorter window for testing
	security.SetRateLimitWindow(5 * time.Second)
	security.SetMaxRequests(2) // Allow only 2 requests per window

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Api-Key", "test:key")

	// First request should succeed
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "First request should succeed")

	// Second request should succeed
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Second request should succeed")

	// Third request should be rate limited
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Third request should be rate limited")

	// Advance time past the window
	mockTime.Advance(5*time.Second + time.Millisecond)

	// Request should succeed again after window expires
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Request after window should succeed")

	// Verify metrics
	assert.Equal(t, int64(3), m.GetSuccessfulRequests(), "Should have 3 successful requests")
	assert.Equal(t, int64(0), m.GetFailedRequests(), "Should have no failed requests")
	assert.Equal(t, int64(1), m.GetRateLimitedRequests(), "Should have 1 rate limited request")
}

func TestCORS(t *testing.T) {
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, _, m, _ := setupTestRouter(t, cfg)

	// Test CORS preflight request
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code, "CORS preflight request should succeed")

	// Test CORS actual request
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("X-Api-Key", "test:key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "CORS actual request should succeed")

	// Verify CORS headers
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, X-Api-Key", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))

	// Verify metrics
	assert.Equal(t, int64(1), m.GetSuccessfulRequests(), "Should have 1 successful request")
	assert.Equal(t, int64(0), m.GetFailedRequests(), "Should have no failed requests")
}

func TestSecurityHeaders(t *testing.T) {
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, _, m, _ := setupTestRouter(t, cfg)

	// Make request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Api-Key", "test:key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))

	// Verify metrics
	assert.Equal(t, int64(1), m.GetSuccessfulRequests(), "Should have 1 successful request")
	assert.Equal(t, int64(0), m.GetFailedRequests(), "Should have no failed requests")
}

func TestMetrics(t *testing.T) {
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}

	// Setup test router with metrics
	router, _, m, mockTime := setupTestRouter(t, cfg)

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Api-Key", "test:key")

	// Make first request (should succeed)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "First request should succeed")

	// Make second request (should succeed)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Second request should succeed")

	// Make third request (should be rate limited)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Third request should be rate limited")

	// Advance time past the rate limit window
	mockTime.Advance(5*time.Second + time.Millisecond)

	// Make another request (should succeed)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Request after window should succeed")

	// Verify metrics
	assert.Equal(t, int64(3), m.GetSuccessfulRequests(), "Should have 3 successful requests")
	assert.Equal(t, int64(0), m.GetFailedRequests(), "Should have no failed requests")
	assert.Equal(t, int64(1), m.GetRateLimitedRequests(), "Should have 1 rate limited request")
}
