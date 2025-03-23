// Package api_test implements tests for the API package.
package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModule tests that the API module provides all necessary dependencies
func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockSearchManager.EXPECT().Close().Return(nil).AnyTimes()
	mockCfg := configtest.NewMockConfig()

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.SearchManager { return mockSearchManager },
			func() config.Interface { return mockCfg },
		),
		api.Module,
	)

	require.NoError(t, app.Err())
}

// TestServerConfiguration tests the server configuration
func TestServerConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockSearchManager.EXPECT().Close().Return(nil).AnyTimes()
	mockCfg := configtest.NewMockConfig()

	tests := []struct {
		name         string
		configPort   int
		expectedPort int
		readTimeout  int
		writeTimeout int
		idleTimeout  int
	}{
		{
			name:         "default_configuration",
			configPort:   8080,
			expectedPort: 8080,
			readTimeout:  10,
			writeTimeout: 30,
			idleTimeout:  60,
		},
		{
			name:         "custom_configuration",
			configPort:   8083,
			expectedPort: 8083,
			readTimeout:  15,
			writeTimeout: 45,
			idleTimeout:  90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCfg.WithServerConfig(&config.ServerConfig{
				Address:      fmt.Sprintf(":%d", tt.configPort),
				ReadTimeout:  time.Duration(tt.readTimeout) * time.Second,
				WriteTimeout: time.Duration(tt.writeTimeout) * time.Second,
				IdleTimeout:  time.Duration(tt.idleTimeout) * time.Second,
			})

			var server *http.Server
			app := fxtest.New(t,
				fx.Provide(
					func() logger.Interface { return mockLogger },
					func() api.SearchManager { return mockSearchManager },
					func() config.Interface { return mockCfg },
				),
				api.Module,
				fx.Populate(&server),
			)

			require.NoError(t, app.Start(t.Context()))
			defer app.Stop(t.Context())

			// Verify server configuration
			assert.Equal(t, fmt.Sprintf(":%d", tt.expectedPort), server.Addr)
			assert.Equal(t, time.Duration(tt.readTimeout)*time.Second, server.ReadTimeout)
			assert.Equal(t, time.Duration(tt.writeTimeout)*time.Second, server.WriteTimeout)
			assert.Equal(t, time.Duration(tt.idleTimeout)*time.Second, server.IdleTimeout)
		})
	}
}

// TestSearchHandler tests the search handler functionality
func TestSearchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockSearchManager.EXPECT().Close().Return(nil).AnyTimes()
	mockCfg := configtest.NewMockConfig()

	tests := []struct {
		name           string
		request        api.SearchRequest
		setupMocks     func()
		expectedStatus int
		expectedBody   api.SearchResponse
	}{
		{
			name: "successful_search",
			request: api.SearchRequest{
				Query: "test query",
				Index: "articles",
				Size:  10,
			},
			setupMocks: func() {
				mockSearchManager.EXPECT().
					Search(gomock.Any(), "articles", gomock.Any()).
					Return([]any{"result1", "result2"}, nil)
				mockSearchManager.EXPECT().
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
			name: "invalid_request",
			request: api.SearchRequest{
				Query: "", // Empty query
				Index: "articles",
				Size:  10,
			},
			setupMocks: func() {
				// Set up expectations for any calls that might happen
				mockSearchManager.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
				mockSearchManager.EXPECT().
					Count(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   api.SearchResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock controller for each test case
			ctrl.Finish()
			ctrl = gomock.NewController(t)
			mockLogger = logger.NewMockInterface(ctrl)
			mockSearchManager = api.NewMockSearchManager(ctrl)
			mockSearchManager.EXPECT().Close().Return(nil).AnyTimes()
			mockCfg = configtest.NewMockConfig()

			tt.setupMocks()

			// Create router
			router, _ := api.SetupRouter(mockLogger, mockSearchManager, mockCfg)

			// Create test request
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Handle request
			router.ServeHTTP(w, req)

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

// TestHealthHandler tests the health check handler
func TestHealthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockSearchManager.EXPECT().Close().Return(nil).AnyTimes()
	mockCfg := configtest.NewMockConfig()

	// Create router
	router, _ := api.SetupRouter(mockLogger, mockSearchManager, mockCfg)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Handle request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}
