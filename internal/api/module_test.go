// Package api_test implements tests for the API package.
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
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

	// Configure mock config with security settings
	serverConfig := &config.ServerConfig{
		Address: ":0", // Use random port
	}
	serverConfig.Security.Enabled = true
	serverConfig.Security.APIKey = "test-api-key"
	serverConfig.Security.RateLimit = 100
	serverConfig.Security.CORS.Enabled = true
	serverConfig.Security.CORS.AllowedOrigins = []string{"*"}
	serverConfig.Security.CORS.AllowedMethods = []string{"GET", "POST"}
	serverConfig.Security.CORS.AllowedHeaders = []string{"Content-Type", "X-Api-Key"}
	serverConfig.Security.CORS.MaxAge = 3600

	mockConfig := configtest.NewMockConfig().
		WithServerConfig(serverConfig).
		WithAppConfig(&config.AppConfig{
			Environment: "test",
			Name:        "test-app",
			Version:     "1.0.0",
			Debug:       true,
		}).
		WithLogConfig(&config.LogConfig{
			Level: "info",
			Debug: true,
		}).
		WithElasticsearchConfig(&config.ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "test_index",
		}).
		WithSources([]config.Source{
			{
				Name:         "test-source",
				URL:          "http://test.com",
				RateLimit:    time.Second,
				MaxDepth:     3,
				ArticleIndex: "articles",
				Index:        "content",
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{},
				},
			},
		})

	var server *http.Server
	mockSecurity := NewMockSecurityMiddleware(serverConfig, mockLogger)
	mockSecurity.On("Cleanup", gomock.Any()).Return()
	mockSecurity.On("WaitCleanup").Return()
	mockSecurity.On("Middleware").Return(func(c *gin.Context) { c.Next() })

	app := fxtest.New(t,
		fx.Supply(
			fx.Annotate(mockLogger, fx.As(new(common.Logger))),
			fx.Annotate(mockConfig, fx.As(new(common.Config))),
		),
		fx.Provide(
			t.Context,
			func() api.SearchManager { return mockSearchMgr },
			func(
				log common.Logger,
				searchManager api.SearchManager,
				cfg common.Config,
				lc fx.Lifecycle,
			) (*http.Server, middleware.SecurityMiddlewareInterface) {
				// Create router and security middleware
				router, security := api.SetupRouter(log, searchManager, cfg)

				// Create server
				server := &http.Server{
					Addr:              cfg.GetServerConfig().Address,
					Handler:           router,
					ReadHeaderTimeout: 10 * time.Second, // Use a reasonable default for tests
				}

				// Register lifecycle hooks
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						log.Info("StartHTTPServer function called")
						log.Info("Server configuration", "address", server.Addr)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						if err := searchManager.Close(); err != nil {
							return fmt.Errorf("search manager close failed: %w", err)
						}
						return nil
					},
				})

				return server, security
			},
		),
		fx.Decorate(
			func(server *http.Server, security middleware.SecurityMiddlewareInterface) (*http.Server, middleware.SecurityMiddlewareInterface) {
				return server, mockSecurity
			},
		),
		fx.Populate(&server),
		fx.StartTimeout(5*time.Second),
		fx.StopTimeout(10*time.Second),
	)

	return &testSetup{
		ctrl:          ctrl,
		mockLogger:    mockLogger,
		mockConfig:    mockConfig,
		mockSearchMgr: mockSearchMgr,
		app:           app,
		server:        server,
	}
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
	ts.mockLogger.EXPECT().Info(
		"Gin request",
		"method", "POST",
		"path", "/search",
		"status", 200,
		"latency", gomock.Any(),
		"client_ip", "192.0.2.1",
		"query", "test query",
		"error", "",
	).Times(1)

	ts.mockSearchMgr.EXPECT().
		Search(gomock.Any(), "articles", gomock.Any()).
		Return([]any{"result1", "result2"}, nil)
	ts.mockSearchMgr.EXPECT().
		Count(gomock.Any(), "articles", gomock.Any()).
		Return(int64(2), nil)
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

	tests := []struct {
		name           string
		request        api.SearchRequest
		setupMocks     func()
		apiKey         string
		expectedStatus int
		expectedBody   api.SearchResponse
	}{
		{
			name: "successful search with valid API key",
			request: api.SearchRequest{
				Query: "test query",
				Index: "articles",
				Size:  10,
			},
			apiKey: "test-api-key",
			setupMocks: func() {
				ts.mockLogger.EXPECT().Info(
					"Gin request",
					"method", "POST",
					"path", "/search",
					"status", 200,
					"latency", gomock.Any(),
					"client_ip", "192.0.2.1",
					"query", "test query",
					"error", "",
				).Times(1)

				ts.mockSearchMgr.EXPECT().
					Search(gomock.Any(), "articles", gomock.Any()).
					Return([]any{"result1", "result2"}, nil)
				ts.mockSearchMgr.EXPECT().
					Count(gomock.Any(), "articles", gomock.Any()).
					Return(int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: api.SearchResponse{
				Results: []any{"result1", "result2"},
				Total:   2,
			},
		},
		{
			name: "unauthorized - missing API key",
			request: api.SearchRequest{
				Query: "test query",
				Index: "articles",
				Size:  10,
			},
			apiKey: "",
			setupMocks: func() {
				ts.mockLogger.EXPECT().Info(
					"Gin request",
					"method", "POST",
					"path", "/search",
					"status", 401,
					"latency", gomock.Any(),
					"client_ip", "192.0.2.1",
					"query", "",
					"error", "Unauthorized",
				).Times(1)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   api.SearchResponse{},
		},
		{
			name: "unauthorized - invalid API key",
			request: api.SearchRequest{
				Query: "test query",
				Index: "articles",
				Size:  10,
			},
			apiKey: "invalid-key",
			setupMocks: func() {
				ts.mockLogger.EXPECT().Info(
					"Gin request",
					"method", "POST",
					"path", "/search",
					"status", 401,
					"latency", gomock.Any(),
					"client_ip", "192.0.2.1",
					"query", "",
					"error", "Unauthorized",
				).Times(1)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   api.SearchResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.setupMocks()

			ts.app.RequireStart()
			t.Cleanup(func() { ts.app.RequireStop() })

			// Create test request with body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
			req.RemoteAddr = "192.0.2.1"
			req.Header.Set("Content-Type", "application/json")
			if tt.apiKey != "" {
				req.Header.Set("X-Api-Key", tt.apiKey)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			ts.server.Handler.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response api.SearchResponse
				jsonErr := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, jsonErr)
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

	ts.mockLogger.EXPECT().Info(
		"Gin request",
		"method", "GET",
		"path", "/health",
		"status", 200,
		"latency", gomock.Any(),
		"client_ip", "192.0.2.1",
		"query", "",
		"error", "",
	).Times(1)

	ts.app.RequireStart()
	t.Cleanup(func() { ts.app.RequireStop() })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "192.0.2.1"
	w := httptest.NewRecorder()

	ts.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

// TestSetupLifecycle verifies the lifecycle hooks for the API server
func TestSetupLifecycle(t *testing.T) {
	t.Parallel()
	ts := setupTest(t)
	t.Cleanup(func() { ts.ctrl.Finish() })

	// Set up mock expectations before starting the app
	ts.mockLogger.EXPECT().Info(
		"StartHTTPServer function called",
		"address", ":0",
	).Times(1)
	ts.mockLogger.EXPECT().Info(
		"Gin request",
		"method", "POST",
		"path", "/search",
		"status", 200,
		"latency", gomock.Any(),
		"client_ip", "192.0.2.1",
		"query", "test query",
		"error", "",
	).Times(1)

	ts.mockSearchMgr.EXPECT().
		Search(gomock.Any(), "articles", gomock.Any()).
		Return([]any{"result1", "result2"}, nil)
	ts.mockSearchMgr.EXPECT().
		Count(gomock.Any(), "articles", gomock.Any()).
		Return(int64(2), nil)
	ts.mockSearchMgr.EXPECT().Close().Return(nil).AnyTimes() // Allow multiple calls

	// Start the app
	ts.app.RequireStart()

	// Stop the app
	ts.app.RequireStop()
	require.NoError(t, ts.app.Err())

	// Verify all expectations
}
