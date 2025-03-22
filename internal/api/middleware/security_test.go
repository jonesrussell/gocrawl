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
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// mockLogger implements logger.Interface for testing
type mockLogger struct {
	logger.Interface
}

func (m *mockLogger) Debug(msg string, fields ...any)   {}
func (m *mockLogger) Error(msg string, fields ...any)   {}
func (m *mockLogger) Info(msg string, fields ...any)    {}
func (m *mockLogger) Warn(msg string, fields ...any)    {}
func (m *mockLogger) Fatal(msg string, fields ...any)   {}
func (m *mockLogger) Printf(format string, args ...any) {}
func (m *mockLogger) Errorf(format string, args ...any) {}
func (m *mockLogger) Sync() error                       { return nil }

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

func setupTestRouter(t *testing.T, securityConfig *config.ServerConfig) (*gin.Engine, *middleware.SecurityMiddleware) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	securityMiddleware := middleware.NewSecurityMiddleware(securityConfig, &mockLogger{})
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
			cfg := &config.ServerConfig{
				Security: struct {
					Enabled   bool   `yaml:"enabled"`
					APIKey    string `yaml:"api_key"`
					RateLimit int    `yaml:"rate_limit"`
					CORS      struct {
						Enabled        bool     `yaml:"enabled"`
						AllowedOrigins []string `yaml:"allowed_origins"`
						AllowedMethods []string `yaml:"allowed_methods"`
						AllowedHeaders []string `yaml:"allowed_headers"`
						MaxAge         int      `yaml:"max_age"`
					} `yaml:"cors"`
					TLS struct {
						Enabled     bool   `yaml:"enabled"`
						Certificate string `yaml:"certificate"`
						Key         string `yaml:"key"`
					} `yaml:"tls"`
				}{
					Enabled: true,
					APIKey:  tt.apiKey,
				},
			}
			router, _ := setupTestRouter(t, cfg)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/test", nil)
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
	cfg := &config.ServerConfig{
		Security: struct {
			Enabled   bool   `yaml:"enabled"`
			APIKey    string `yaml:"api_key"`
			RateLimit int    `yaml:"rate_limit"`
			CORS      struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			} `yaml:"cors"`
			TLS struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			} `yaml:"tls"`
		}{
			Enabled:   true,
			APIKey:    "test-key",
			RateLimit: 2, // 2 requests per 5 seconds
		},
	}
	router, securityMiddleware := setupTestRouter(t, cfg)

	// Set a shorter window for testing
	securityMiddleware.SetRateLimitWindow(5 * time.Second)

	// Set up mock time provider
	mockTime := &mockTimeProvider{currentTime: time.Now()}
	securityMiddleware.SetTimeProvider(mockTime)

	// Start cleanup goroutine with a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start cleanup in a goroutine
	cleanupDone := make(chan struct{})
	go func() {
		securityMiddleware.Cleanup(ctx)
		close(cleanupDone)
	}()

	// Make requests from the same IP
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Api-Key", "test-key")
	req.RemoteAddr = "127.0.0.1"

	// First request should succeed
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should succeed
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Advance time by 5 seconds
	mockTime.Advance(5 * time.Second)

	// Request should succeed again
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Cancel the context to stop the cleanup goroutine
	cancel()

	// Wait for cleanup to finish
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
		origin         string
		method         string
		allowedOrigins []string
		expectedStatus int
	}{
		{
			name:           "allowed origin",
			origin:         "http://example.com",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "disallowed origin",
			origin:         "http://malicious.com",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "preflight request",
			origin:         "http://example.com",
			method:         http.MethodOptions,
			allowedOrigins: []string{"http://example.com"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ServerConfig{
				Security: struct {
					Enabled   bool   `yaml:"enabled"`
					APIKey    string `yaml:"api_key"`
					RateLimit int    `yaml:"rate_limit"`
					CORS      struct {
						Enabled        bool     `yaml:"enabled"`
						AllowedOrigins []string `yaml:"allowed_origins"`
						AllowedMethods []string `yaml:"allowed_methods"`
						AllowedHeaders []string `yaml:"allowed_headers"`
						MaxAge         int      `yaml:"max_age"`
					} `yaml:"cors"`
					TLS struct {
						Enabled     bool   `yaml:"enabled"`
						Certificate string `yaml:"certificate"`
						Key         string `yaml:"key"`
					} `yaml:"tls"`
				}{
					Enabled: true,
					CORS: struct {
						Enabled        bool     `yaml:"enabled"`
						AllowedOrigins []string `yaml:"allowed_origins"`
						AllowedMethods []string `yaml:"allowed_methods"`
						AllowedHeaders []string `yaml:"allowed_headers"`
						MaxAge         int      `yaml:"max_age"`
					}{
						Enabled:        true,
						AllowedOrigins: tt.allowedOrigins,
						AllowedMethods: []string{http.MethodPost},
						AllowedHeaders: []string{"Content-Type", "X-Api-Key"},
						MaxAge:         86400,
					},
				},
			}
			router, _ := setupTestRouter(t, cfg)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, "/test", nil)
			require.NoError(t, err)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", http.MethodPost)
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
	cfg := &config.ServerConfig{
		Security: struct {
			Enabled   bool   `yaml:"enabled"`
			APIKey    string `yaml:"api_key"`
			RateLimit int    `yaml:"rate_limit"`
			CORS      struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			} `yaml:"cors"`
			TLS struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			} `yaml:"tls"`
		}{
			Enabled: true,
		},
	}
	router, _ := setupTestRouter(t, cfg)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/test", nil)
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
