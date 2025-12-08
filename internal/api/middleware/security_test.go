package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	loggerMock "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"go.uber.org/mock/gomock"
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
) (*gin.Engine, *middleware.SecurityMiddleware, *mockTimeProvider) {
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

	return router, security, mockTime
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
			name: "handles OPTIONS request",
			config: &server.Config{
				Address: ":8080",
			},
			origin:         "http://test.com",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "handles request without origin",
			config: &server.Config{
				Address: ":8080",
			},
			origin:         "",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router, _, _ := setupTestRouter(t, tt.config)

			req := httptest.NewRequest(tt.method, "/test", http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.origin != "" {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type, Authorization, X-API-Key", w.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			}
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
			router, _, _ := setupTestRouter(t, tt.config)

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

	// Setup test router with security middleware
	cfg := &server.Config{
		SecurityEnabled: true,
		APIKey:          "test-key",
		Address:         ":8080",
	}
	router, security, mockTime := setupTestRouter(t, cfg)

	// Set a very short window for testing
	security.SetRateLimitWindow(100 * time.Millisecond)
	security.SetMaxRequests(2)

	// First request should succeed
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second request should succeed
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Third request should be rate limited
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Wait for rate limit window to expire
	mockTime.Advance(200 * time.Millisecond)

	// Request should succeed again
	req = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Api-Key", "test-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
