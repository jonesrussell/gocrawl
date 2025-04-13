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
	"github.com/jonesrussell/gocrawl/internal/metrics"
	loggerMock "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
)

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
func setupTestRouter(
	t *testing.T,
	cfg *server.Config,
) (*gin.Engine, *middleware.SecurityMiddleware, *metrics.Metrics, *mockTimeProvider) {
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
			t.Parallel()
			router, _, _, _ := setupTestRouter(t, tt.config)

			req := httptest.NewRequest(tt.method, "/test", http.NoBody)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
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
			t.Parallel()
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
