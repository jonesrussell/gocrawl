package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
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

var (
	setupOnce sync.Once
	portMutex sync.Mutex
	nextPort  = 8081
)

func init() {
	// Set Gin to test mode once at startup
	gin.SetMode(gin.TestMode)
}

// getTestPort returns a unique port for testing
func getTestPort() string {
	portMutex.Lock()
	defer portMutex.Unlock()
	port := nextPort
	nextPort++
	return fmt.Sprintf(":%d", port)
}

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

// setupTestRouter creates a new test router with security middleware
func setupTestRouter(t *testing.T, cfg *server.Config) (*gin.Engine, *middleware.SecurityMiddleware, *metrics.Metrics, *mockTimeProvider) {
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
	mockTime := &mockTimeProvider{}
	security.SetTimeProvider(mockTime)

	router := gin.New()
	router.Use(security.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	metrics := security.GetMetrics()
	return router, security, metrics, mockTime
}

func TestAPIKeyAuthentication(t *testing.T) {
	t.Parallel()
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, security, m, _ := setupTestRouter(t, cfg)
	t.Cleanup(func() {
		security.ResetRateLimiter()
		m.ResetMetrics()
	})

	// Test missing API key
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	t.Parallel()
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, security, m, mockTime := setupTestRouter(t, cfg)
	t.Cleanup(func() {
		security.ResetRateLimiter()
		m.ResetMetrics()
	})
	security.SetRateLimitWindow(5 * time.Second)
	security.SetMaxRequests(2) // Allow only 2 requests per window
	m.ResetMetrics()           // Reset metrics before testing

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	t.Parallel()

	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	router, security, metrics, _ := setupTestRouter(t, cfg)
	t.Cleanup(func() {
		security.ResetRateLimiter()
		metrics.ResetMetrics()
	})

	// Test CORS preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", http.NoBody)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Test actual CORS request
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("X-Api-Key", "test:key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Verify metrics
	assert.Equal(t, int64(2), metrics.GetSuccessfulRequests())
	assert.Equal(t, int64(0), metrics.GetFailedRequests())
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	// Create test configuration
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test:key",
		Address:         ":8080",
	}

	// Setup test router
	router, security, m, _ := setupTestRouter(t, cfg)
	t.Cleanup(func() {
		security.ResetRateLimiter()
		m.ResetMetrics()
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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
	t.Parallel()
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
	router, security, m, mockTime := setupTestRouter(t, cfg)
	t.Cleanup(func() {
		security.ResetRateLimiter()
		m.ResetMetrics()
	})
	security.SetRateLimitWindow(5 * time.Second)
	security.SetMaxRequests(2) // Allow only 2 requests per window
	m.ResetMetrics()           // Reset metrics before testing

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
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

func TestSecurityMiddleware_HandleCORS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *server.Config
		origin         string
		method         string
		expectedStatus int
	}{
		{
			name: "test environment allows any origin",
			config: &server.Config{
				Address: ":8080",
			},
			origin:         "http://test.com",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name: "production environment allows only example.com",
			config: &server.Config{
				Address: ":9090",
			},
			origin:         "https://example.com",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name: "production environment rejects non-example.com",
			config: &server.Config{
				Address: ":9090",
			},
			origin:         "https://other.com",
			method:         http.MethodGet,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _, _, _ := setupTestRouter(t, tt.config)

			req := httptest.NewRequest(tt.method, "/test", http.NoBody)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_OptionsRequest(t *testing.T) {
	t.Parallel()

	router, _, _, _ := setupTestRouter(t, &server.Config{
		Address: ":8080",
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", http.NoBody)
	req.Header.Set("Origin", "http://test.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurityMiddleware_APIAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *server.Config
		apiKey         string
		expectedStatus int
	}{
		{
			name: "missing API key",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "test-key",
			},
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid API key",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "test-key",
			},
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid API key",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "test-key",
			},
			apiKey:         "test-key",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _, _, _ := setupTestRouter(t, tt.config)

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_RateLimit(t *testing.T) {
	t.Parallel()

	router, security, metrics, mockTime := setupTestRouter(t, &server.Config{
		Address: ":8080",
	})

	// Set a very short window for testing
	security.SetRateLimitWindow(100 * time.Millisecond)
	security.SetMaxRequests(2)

	// First request should succeed
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should succeed
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Verify metrics
	assert.Equal(t, int64(2), metrics.GetSuccessfulRequests())
	assert.Equal(t, int64(1), metrics.GetRateLimitedRequests())

	// Wait for rate limit window to expire
	mockTime.Advance(200 * time.Millisecond)

	// Request should succeed again
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
