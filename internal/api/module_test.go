// Package api_test implements tests for the API package.
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

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
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// testSetup contains common test dependencies
type testSetup struct {
	t             *testing.T
	ctrl          *gomock.Controller
	mockLogger    *logger.MockInterface
	mockConfig    *configtest.MockConfig
	mockSearchMgr *api.MockSearchManager
	app           *fxtest.App
	server        *http.Server
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
		Address:      ":0", // Use random port
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
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
	mockSearchMgr.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interface{}{}, nil).AnyTimes()
	mockSearchMgr.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), nil).AnyTimes()

	// Create test setup
	ts := &testSetup{
		t:             t,
		ctrl:          ctrl,
		mockLogger:    mockLogger,
		mockSearchMgr: mockSearchMgr,
		mockConfig:    mockConfig,
	}

	// Set up lifecycle hooks
	ts.app = fxtest.New(t,
		fx.Supply(mockLogger, mockConfig),
		fx.Provide(
			func() api.SearchManager { return mockSearchMgr },
			func() *http.Server {
				server := &http.Server{
					Addr:              serverConfig.Address,
					ReadHeaderTimeout: 10 * time.Second,
				}
				return server
			},
			func() middleware.SecurityMiddlewareInterface {
				return middleware.NewSecurityMiddleware(serverConfig, mockLogger)
			},
		),
		fx.Decorate(
			func(server *http.Server) *http.Server {
				router, _ := api.SetupRouter(mockLogger, mockSearchMgr, mockConfig)
				server.Handler = router
				return server
			},
			func(security middleware.SecurityMiddlewareInterface) middleware.SecurityMiddlewareInterface {
				return security
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, server *http.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							mockLogger.Error("Server error", "error", err)
						}
					}()
					// Wait for server to be ready
					time.Sleep(100 * time.Millisecond)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
		}),
	)

	return ts
}

// TestModuleConstruction verifies that the API module can be constructed with all dependencies
func TestModuleConstruction(t *testing.T) {
	t.Parallel()
	ts := setupTest(t)
	t.Cleanup(func() { ts.ctrl.Finish() })

	// Set up mock expectations before starting the app
	ts.mockLogger.EXPECT().Info(
		"StartHTTPServer function called",
		"address", ":0",
	).Times(1)
	ts.mockSearchMgr.EXPECT().Close().Return(nil).AnyTimes() // Allow multiple calls

	// Start the app
	ts.app.RequireStart()

	// Stop the app
	ts.app.RequireStop()
	require.NoError(t, ts.app.Err())

	// Verify all expectations
}

// TestSearchHandler verifies the search endpoint functionality with security
func TestSearchHandler(t *testing.T) {
	t.Parallel()
	ts := setupTest(t)
	t.Cleanup(func() { ts.ctrl.Finish() })

	// Set up mock expectations
	ts.mockSearchMgr.EXPECT().
		Search(gomock.Any(), "articles", gomock.Any()).
		Return([]interface{}{}, nil)
	ts.mockSearchMgr.EXPECT().
		Count(gomock.Any(), "articles", gomock.Any()).
		Return(int64(0), nil)

	// Start the app
	ts.app.RequireStart()
	t.Cleanup(func() { ts.app.RequireStop() })

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Test cases
	tests := []struct {
		name           string
		request        api.SearchRequest
		apiKey         string
		expectedStatus int
		expectedBody   api.SearchResponse
	}{
		{
			name: "successful search with valid API key",
			request: api.SearchRequest{
				Query: "articles",
			},
			apiKey:         "test-api-key",
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
			// Create test request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP client
			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			// Create request
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/search", ts.server.Addr), bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}

			// Execute request
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedStatus == http.StatusOK {
				var response api.SearchResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

// TestHealthHandler verifies the health check endpoint functionality
func TestHealthHandler(t *testing.T) {
	t.Parallel()
	ts := setupTest(t)
	t.Cleanup(func() { ts.ctrl.Finish() })

	// Start the app
	ts.app.RequireStart()
	t.Cleanup(func() { ts.app.RequireStop() })

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/health", ts.server.Addr), nil)
	require.NoError(t, err)

	// Execute request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "OK", string(body))
}
