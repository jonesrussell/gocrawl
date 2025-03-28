// Package httpd_test implements tests for the HTTP server command.
package httpd_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// setupTestDependencies creates and configures all test dependencies
func setupTestDependencies() (
	*testutils.MockLogger,
	*testutils.MockConfig,
	*httpdTestStorage,
	*testutils.MockSearchManager,
	*testutils.MockSecurityMiddleware,
) {
	mockLogger := &testutils.MockLogger{}
	mockCfg := &testutils.MockConfig{}
	mockStore := &httpdTestStorage{}
	mockSearch := &testutils.MockSearchManager{}
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Set up mock config expectations
	mockCfg.On("GetAppConfig").Return(&config.AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockCfg.On("GetLogConfig").Return(&config.LogConfig{
		Level: "debug",
		Debug: true,
	})
	mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	})
	mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
	mockCfg.On("GetSources").Return([]config.Source{}, nil)
	mockCfg.On("GetCommand").Return("test")

	return mockLogger, mockCfg, mockStore, mockSearch, mockSecurity
}

// createTestModule creates a test-specific module with all required dependencies
func createTestModule(mockStore *httpdTestStorage) fx.Option {
	return fx.Module("test",
		common.Module,
		fx.Supply(
			mockStore,
		),
		fx.Provide(
			context.Background,
			fx.Annotate(
				func() storagetypes.Interface { return mockStore },
				fx.As(new(storagetypes.Interface)),
			),
			createServerAndSecurity,
		),
	)
}

// createServerAndSecurity creates the HTTP server and security middleware
func createServerAndSecurity(
	log common.Logger,
	cfg common.Config,
	lc fx.Lifecycle,
) (*http.Server, middleware.SecurityMiddlewareInterface) {
	// Create router and security middleware
	router, security := api.SetupRouter(log, nil, cfg)

	// Create server
	server := &http.Server{
		Addr:              cfg.GetServerConfig().Address,
		Handler:           router,
		ReadHeaderTimeout: api.ReadHeaderTimeout,
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Create a channel to signal when the server is ready
			ready := make(chan struct{})
			go func() {
				// Start the server
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("Server error", "error", err)
				}
			}()

			// Create a timeout context for health check
			healthCtx, cancel := context.WithTimeout(ctx, api.HealthCheckTimeout)
			defer cancel()

			// Create a ticker for health check attempts
			ticker := time.NewTicker(api.HealthCheckInterval)
			defer ticker.Stop()

			// Try to connect to the health endpoint until successful or timeout
			for {
				select {
				case <-healthCtx.Done():
					return fmt.Errorf("server failed to become healthy within %v", api.HealthCheckTimeout)
				case <-ticker.C:
					// Create a temporary client for health check
					client := &http.Client{
						Timeout: api.HealthCheckInterval,
					}

					// Try to connect to the health endpoint
					resp, err := client.Get(fmt.Sprintf("http://%s/health", server.Addr))
					if err != nil {
						continue // Server not ready yet
					}
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						close(ready)
						return nil
					}
				}
			}
		},
		OnStop: func(ctx context.Context) error {
			// Create a timeout context for shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, api.ShutdownTimeout)
			defer cancel()

			// Shutdown server
			if err := server.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("server shutdown failed: %w", err)
			}

			// Cleanup security middleware
			security.Cleanup(ctx)
			security.WaitCleanup()

			return nil
		},
	})

	return server, security
}

// TestModuleProvides tests that the HTTPD module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	t.Parallel()

	_, _, mockStore, _, _ := setupTestDependencies()
	testModule := createTestModule(mockStore)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule,
	)

	require.NoError(t, app.Err())
}

// httpdTestStorage implements storagetypes.Interface for testing
type httpdTestStorage struct {
	storagetypes.Interface
}
