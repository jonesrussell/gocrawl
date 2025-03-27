// Package api_test implements tests for the API package.
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// testSetup contains common test dependencies
type testSetup struct {
	t             *testing.T
	ctrl          *gomock.Controller
	mockLogger    *logger.MockInterface
	mockConfig    *configtest.MockConfig
	mockSearchMgr *api.MockSearchManager
	router        *gin.Engine
}

// MockSecurityMiddleware implements SecurityMiddleware interface for testing
type MockSecurityMiddleware struct {
	mock.Mock
}

// NewMockSecurityMiddleware creates a new mock security middleware
func NewMockSecurityMiddleware(cfg *config.ServerConfig, log logger.Interface) *MockSecurityMiddleware {
	return &MockSecurityMiddleware{}
}

func (m *MockSecurityMiddleware) Cleanup(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockSecurityMiddleware) WaitCleanup() {
	m.Called()
}

func (m *MockSecurityMiddleware) Middleware() gin.HandlerFunc {
	args := m.Called()
	if fn := args.Get(0); fn != nil {
		return fn.(gin.HandlerFunc)
	}
	return func(c *gin.Context) { c.Next() }
}

// Ensure MockSecurityMiddleware implements SecurityMiddlewareInterface
var _ middleware.SecurityMiddlewareInterface = (*MockSecurityMiddleware)(nil)

// setupTest creates a new test environment with mocked dependencies
func setupTest(t *testing.T) *testSetup {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockInterface(ctrl)
	mockSearchMgr := api.NewMockSearchManager(ctrl)
	mockConfig := configtest.NewMockConfig()

	// Set up server config
	serverConfig := &config.ServerConfig{
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
			Enabled:   true, // Enable security for proper testing
			APIKey:    "test-key",
			RateLimit: 100,
			CORS: struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			}{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Authorization", "X-API-Key"},
				MaxAge:         86400,
			},
		},
	}

	// Set up mock config
	mockConfig.WithServerConfig(serverConfig)

	// Set up mock expectations
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create router
	router, _ := api.SetupRouter(mockLogger, mockSearchMgr, mockConfig)

	// Create test setup
	ts := &testSetup{
		t:             t,
		ctrl:          ctrl,
		mockLogger:    mockLogger,
		mockSearchMgr: mockSearchMgr,
		mockConfig:    mockConfig,
		router:        router,
	}

	return ts
}

// TestSearchHandler verifies the search endpoint functionality with security
func TestSearchHandler(t *testing.T) {
	ts := setupTest(t)
	defer ts.ctrl.Finish()

	// Set up base mock expectations for successful case
	expectedQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "articles",
			},
		},
		"size": 0,
	}
	countQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "articles",
			},
		},
	}

	// Test cases
	tests := []struct {
		name           string
		request        api.SearchRequest
		apiKey         string
		setupMocks     func()
		expectedStatus int
		expectedBody   api.SearchResponse
	}{
		{
			name: "successful search with valid API key",
			request: api.SearchRequest{
				Query: "articles",
				Size:  0,
			},
			apiKey: "test-key",
			setupMocks: func() {
				ts.mockSearchMgr.EXPECT().
					Search(gomock.Any(), "", gomock.Eq(expectedQuery)).
					Return([]interface{}{}, nil).
					Times(1)
				ts.mockSearchMgr.EXPECT().
					Count(gomock.Any(), "", gomock.Eq(countQuery)).
					Return(int64(0), nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: api.SearchResponse{
				Results: []interface{}{},
				Total:   0,
			},
		},
		{
			name: "unauthorized - missing API key",
			request: api.SearchRequest{
				Query: "articles",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "unauthorized - invalid API key",
			request: api.SearchRequest{
				Query: "articles",
			},
			apiKey:         "invalid-key",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// Create test request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create request and recorder
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/search", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}

			// Execute request
			ts.router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response api.SearchResponse
				err = json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

// TestHealthHandler verifies the health check endpoint functionality
func TestHealthHandler(t *testing.T) {
	ts := setupTest(t)
	defer ts.ctrl.Finish()

	// Create request and recorder
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Api-Key", "test-key") // Add API key for auth

	// Execute request
	ts.router.ServeHTTP(w, req)

	// Read response body
	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"status\":\"ok\"}", string(body))
}
