package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx/fxevent"
)

// mockLogger implements common.Logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) With(fields ...logger.Field) logger.Interface {
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

// mockTimeProvider implements TimeProvider for testing
type mockTimeProvider struct {
	currentTime time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *mockTimeProvider) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

func setupTestRouter(t *testing.T, securityConfig *server.Config) (*gin.Engine, *middleware.SecurityMiddleware) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLog := &mockLogger{}
	mockLog.On("Info", mock.Anything, mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()
	mockLog.On("Fatal", mock.Anything, mock.Anything).Return()
	mockLog.On("With", mock.Anything).Return(mockLog)

	securityMiddleware := middleware.NewSecurityMiddleware(securityConfig, mockLog)
	router.Use(securityMiddleware.Middleware())

	router.POST("/test", func(c *gin.Context) {
		t.Log("Handling test request")
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	return router, securityMiddleware
}

func TestAPIKeyAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		requestKey     string
		expectedStatus int
	}{
		{
			name:           "valid api key",
			apiKey:         "test-key",
			requestKey:     "test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid api key",
			apiKey:         "test-key",
			requestKey:     "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing api key",
			apiKey:         "test-key",
			requestKey:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &server.Config{
				SecurityEnabled: true,
				APIKey:          tt.apiKey,
				Address:         ":8080",
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
			}
			router, _ := setupTestRouter(t, cfg)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/test", http.NoBody)
			require.NoError(t, err)

			if tt.requestKey != "" {
				req.Header.Set("X-Api-Key", tt.requestKey)
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRateLimiting(t *testing.T) {
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test-key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
	router, securityMiddleware := setupTestRouter(t, cfg)

	// Set a shorter window for testing
	securityMiddleware.SetRateLimitWindow(5 * time.Second)

	// Set up mock time provider with a fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mockTime := &mockTimeProvider{currentTime: startTime}
	securityMiddleware.SetTimeProvider(mockTime)

	// Start cleanup goroutine with a context that we can cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Start cleanup in a goroutine
	cleanupDone := make(chan struct{})
	go func() {
		securityMiddleware.Cleanup(ctx)
		close(cleanupDone)
	}()

	// Make requests from the same IP
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/test", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("X-Api-Key", "test-key")
	req.RemoteAddr = "127.0.0.1"

	// First request should succeed
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

	// Advance time by 5 seconds and 1 millisecond to ensure we're past the window
	mockTime.Advance(5*time.Second + time.Millisecond)

	// Give the cleanup goroutine a chance to run
	time.Sleep(100 * time.Millisecond)

	// Request should succeed again after window expires
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Request after window should succeed")

	// Cancel the context to stop the cleanup goroutine
	cancel()

	// Wait for cleanup to finish with a timeout
	select {
	case <-cleanupDone:
		// Cleanup finished successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Cleanup goroutine did not finish within timeout")
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		origin         string
		allowedOrigins []string
		expectedStatus int
	}{
		{
			name:           "allowed origin",
			method:         http.MethodPost,
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "disallowed origin",
			method:         http.MethodPost,
			origin:         "https://evil.com",
			allowedOrigins: []string{"https://example.com"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "preflight request",
			method:         http.MethodOptions,
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &server.Config{
				SecurityEnabled: true,
				APIKey:          "test-key",
				Address:         ":8080",
				ReadTimeout:     15 * time.Second,
				WriteTimeout:    15 * time.Second,
				IdleTimeout:     60 * time.Second,
			}
			router, _ := setupTestRouter(t, cfg)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, "/test", http.NoBody)
			require.NoError(t, err)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			} else {
				req.Header.Set("X-Api-Key", "test-key")
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNoContent {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "X-Api-Key")
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test-key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
	router, _ := setupTestRouter(t, cfg)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/test", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("X-Api-Key", "test-key")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check security headers
	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, expectedValue := range headers {
		assert.Equal(t, expectedValue, w.Header().Get(header))
	}
}

func TestMetrics(t *testing.T) {
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test-key",
		Address:         ":8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
	router, securityMiddleware := setupTestRouter(t, cfg)

	// Set a shorter window for testing
	securityMiddleware.SetRateLimitWindow(5 * time.Second)

	// Set up mock time provider with a fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mockTime := &mockTimeProvider{currentTime: startTime}
	securityMiddleware.SetTimeProvider(mockTime)

	// Start cleanup goroutine with a context that we can cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Start cleanup in a goroutine
	cleanupDone := make(chan struct{})
	go func() {
		securityMiddleware.Cleanup(ctx)
		close(cleanupDone)
	}()

	// Make requests from the same IP
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/test", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("X-Api-Key", "test-key")
	req.RemoteAddr = "127.0.0.1"

	// First request should succeed
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

	// Advance time by 5 seconds and 1 millisecond to ensure we're past the window
	mockTime.Advance(5*time.Second + time.Millisecond)

	// Give the cleanup goroutine a chance to run
	time.Sleep(100 * time.Millisecond)

	// Request should succeed again after window expires
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Request after window should succeed")

	// Cancel the context to stop the cleanup goroutine
	cancel()

	// Wait for cleanup to finish with a timeout
	select {
	case <-cleanupDone:
		// Cleanup finished successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Cleanup goroutine did not finish within timeout")
	}

	// Wait for goroutine to complete
	time.Sleep(common.DefaultTestSleepDuration)

	// Verify metrics
	metrics := metrics.NewMetrics()
	assert.Equal(t, int64(1), metrics.GetProcessedCount())
	assert.Equal(t, int64(0), metrics.GetErrorCount())
}
