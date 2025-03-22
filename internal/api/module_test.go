// Package api_test implements tests for the API package.
package api_test

import (
	"fmt"
	"io"
	"net/http"
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

// waitForServer waits for the server to be ready by retrying the health check
func waitForServer(url string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server not ready after %d retries", maxRetries)
}

// TestModule tests that the API module provides all necessary dependencies
func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSearchManager := api.NewMockSearchManager(ctrl)
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

// TestStartHTTPServer tests the HTTP server startup functionality
func TestStartHTTPServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockCfg := configtest.NewMockConfig()

	tests := []struct {
		name           string
		setupMocks     func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful_search",
			setupMocks: func() {
				mockSearchManager.EXPECT().
					Search(gomock.Any(), "articles", gomock.Any()).
					Return([]any{}, nil)
				mockSearchManager.EXPECT().
					Count(gomock.Any(), "articles", gomock.Any()).
					Return(int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total":0,"results":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			// Configure server to use a fixed port for testing
			mockCfg.WithServerConfig(&config.ServerConfig{
				Address: ":8082", // Use fixed port for testing
			})

			app := fxtest.New(t,
				fx.Provide(
					func() logger.Interface { return mockLogger },
					func() api.SearchManager { return mockSearchManager },
					func() config.Interface { return mockCfg },
				),
				api.Module,
			)

			require.NoError(t, app.Start(t.Context()))
			defer app.Stop(t.Context())

			// Wait for server to be ready
			require.NoError(t, waitForServer("http://localhost:8082/health", 5))

			// Test health endpoint
			resp, err := http.Get("http://localhost:8082/health")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Test search endpoint
			searchBody := `{"query": "test query", "index": "articles", "size": 10}`
			resp, err = http.Post("http://localhost:8082/search", "application/json", strings.NewReader(searchBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expectedBody, string(body))
		})
	}
}

// TestStartHTTPServer_PortConfiguration tests the HTTP server port configuration
func TestStartHTTPServer_PortConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockCfg := configtest.NewMockConfig()

	tests := []struct {
		name           string
		configPort     int
		expectedPort   int
		expectedStatus int
	}{
		{
			name:           "use_config_port",
			configPort:     8083,
			expectedPort:   8083,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCfg.WithServerConfig(&config.ServerConfig{
				Address: fmt.Sprintf(":%d", tt.configPort),
			})

			app := fxtest.New(t,
				fx.Provide(
					func() logger.Interface { return mockLogger },
					func() api.SearchManager { return mockSearchManager },
					func() config.Interface { return mockCfg },
				),
				api.Module,
			)

			require.NoError(t, app.Start(t.Context()))
			defer app.Stop(t.Context())

			// Wait for server to be ready
			require.NoError(t, waitForServer(fmt.Sprintf("http://localhost:%d/health", tt.expectedPort), 5))

			// Test health endpoint
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", tt.expectedPort))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
