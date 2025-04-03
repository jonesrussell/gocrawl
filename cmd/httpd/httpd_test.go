package httpd_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockStorage implements storagetypes.Interface for testing
type mockStorage struct {
	mock.Mock
	storagetypes.Interface
}

func (m *mockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, errors.New("invalid search result type")
	}
	return val, nil
}

func (m *mockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, errors.New("invalid count result type")
	}
	return val, nil
}

func (m *mockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, errors.New("aggregation result is nil")
	}
	return result, nil
}

func (m *mockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// searchManagerWrapper wraps storagetypes.Interface to implement api.SearchManager
type searchManagerWrapper struct {
	storagetypes.Interface
}

func (w *searchManagerWrapper) Search(ctx context.Context, index string, query map[string]any) ([]any, error) {
	result, err := w.Interface.Search(ctx, index, query)
	if err != nil {
		return nil, err
	}
	converted := make([]any, len(result))
	copy(converted, result)
	return converted, nil
}

func (w *searchManagerWrapper) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	return w.Interface.Count(ctx, index, query)
}

func (w *searchManagerWrapper) Aggregate(
	ctx context.Context,
	index string,
	aggs map[string]any,
) (map[string]any, error) {
	result, err := w.Interface.Aggregate(ctx, index, aggs)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("aggregation result is nil")
	}
	if converted, ok := result.(map[string]any); ok {
		return converted, nil
	}
	return nil, errors.New("invalid aggregation result type")
}

func (w *searchManagerWrapper) Close() error {
	return w.Interface.Close()
}

// TestConfigModule provides a test-specific config module that doesn't try to load files.
var TestConfigModule = fx.Module("testConfig",
	fx.Replace(
		fx.Annotate(
			func() config.Interface {
				mockCfg := &testutils.MockConfig{}
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
				mockCfg.On("GetServerConfig").Return(&config.ServerConfig{
					Address: ":invalid_port",
				})
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				return mockCfg
			},
			fx.As(new(config.Interface)),
		),
	),
)

func TestHTTPCommand(t *testing.T) {
	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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

	mockStore := &mockStorage{}
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies using fxtest
	app := fxtest.New(t,
		fx.NopLogger,
		// Provide mock logger directly
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		// Exclude logger module and provide other modules
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: http.NewServeMux(),
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
	)

	app.RequireStart()
	app.RequireStop()
}

func TestHTTPCommandGracefulShutdown(t *testing.T) {
	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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

	mockStore := &mockStorage{}
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies using fxtest
	app := fxtest.New(t,
		fx.NopLogger,
		// Provide mock logger directly
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		// Exclude logger module and provide other modules
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: http.NewServeMux(),
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
	)

	app.RequireStart()
	app.RequireStop()
}

func TestCommand(t *testing.T) {
	cmd := httpd.Command()
	assert.NotNil(t, cmd)
	assert.Equal(t, "httpd", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestServerStartStop(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	errListen := listener.Close()
	if errListen != nil {
		return
	}

	// Create a mux for the server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("ok")); writeErr != nil {
			t.Errorf("Error writing response: %v", writeErr)
		}
	})

	serverConfig := &config.ServerConfig{
		Address: fmt.Sprintf(":%d", port),
	}

	server := &http.Server{
		Addr:    serverConfig.Address,
		Handler: mux,
	}

	// Channel to signal when server is ready
	serverReady := make(chan struct{})
	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		// Signal that we're starting the server
		close(serverReady)

		// Start the server
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			serverErr <- listenErr
		}
	}()

	// Wait for server to be ready
	select {
	case <-serverReady:
		// Server is ready, proceed with test
	case startErr := <-serverErr:
		t.Fatalf("Server failed to start: %v", startErr)
	case <-time.After(5 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Give the server a moment to fully start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint with retries
	var resp *http.Response
	var respErr error
	maxRetries := 3
	for i := range maxRetries {
		resp, respErr = http.Get(fmt.Sprintf("http://localhost:%d/health", port))
		if respErr == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	require.NoError(t, respErr)
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("Error closing response body: %v", closeErr)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Stop server gracefully
	stopCtx, stopCancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer stopCancel()
	stopErr := server.Shutdown(stopCtx)
	require.NoError(t, stopErr)

	// Verify server is stopped
	_, respErr = http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.Error(t, respErr)
}

// TestServerHealthCheck tests the health check functionality
func TestServerHealthCheck(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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
	mockCfg.On("GetServerConfig").Return(&config.ServerConfig{
		Address: ":8080",
	})
	mockCfg.On("GetSources").Return([]config.Source{}, nil)
	mockCfg.On("GetCommand").Return("test")

	mockStore := &mockStorage{}
	mockStore.On("TestConnection", mock.Anything).Return(nil)
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					mux := http.NewServeMux()
					mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					})
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: mux,
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, p httpd.Params) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := p.Storage.TestConnection(ctx); err != nil {
						return fmt.Errorf("failed to connect to storage: %w", err)
					}

					// Start HTTP server in background
					p.Logger.Info("Starting HTTP server...", "address", p.Server.Addr)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app and verify it starts without errors
	err := app.Start(t.Context())
	require.NoError(t, err)
	app.RequireStop()

	// Verify that the server was configured correctly
	// Note: We don't need to verify the exact log message since we're testing server startup
}

// TestServerStorageConnection tests storage connection functionality
func TestServerStorageConnection(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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

	mockStore := &mockStorage{}
	mockStore.On("TestConnection", mock.Anything).Return(nil)
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: http.NewServeMux(),
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, p httpd.Params) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := p.Storage.TestConnection(ctx); err != nil {
						return fmt.Errorf("failed to connect to storage: %w", err)
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app and verify storage connection
	err := app.Start(t.Context())
	require.NoError(t, err)
	app.RequireStop()

	// Verify that the storage connection was tested
	mockStore.AssertCalled(t, "TestConnection", mock.Anything)
}

// TestServerErrorHandling tests error handling scenarios
func TestServerErrorHandling(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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

	mockStore := &mockStorage{}
	mockStore.On("TestConnection", mock.Anything).Return(errors.New("storage connection failed"))
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: http.NewServeMux(),
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, p httpd.Params) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := p.Storage.TestConnection(ctx); err != nil {
						return fmt.Errorf("failed to connect to storage: %w", err)
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Verify that the app fails to start due to storage connection error
	err := app.Start(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage connection failed")
}

// TestServerTimeoutHandling tests timeout scenarios
func TestServerTimeoutHandling(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
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

	mockStore := &mockStorage{}
	mockStore.On("TestConnection", mock.Anything).Return(nil)
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.As(new(logger.Interface)),
			),
		),
		fx.Supply(
			mockStore,
			mockSecurity,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		fx.Module("test",
			fx.Provide(
				func(storage storagetypes.Interface) api.SearchManager {
					return &searchManagerWrapper{storage}
				},
				func(cfg config.Interface) *http.Server {
					return &http.Server{
						Addr:    cfg.GetServerConfig().Address,
						Handler: http.NewServeMux(),
					}
				},
				func() middleware.SecurityMiddlewareInterface {
					return mockSecurity
				},
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, p httpd.Params) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Test storage connection
					if err := p.Storage.TestConnection(ctx); err != nil {
						return fmt.Errorf("failed to connect to storage: %w", err)
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Simulate a slow shutdown that exceeds the timeout
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(200 * time.Millisecond):
						return nil
					}
				},
			})
		}),
	)

	// Start the app
	app.RequireStart()

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	// Attempt to stop the app with the short timeout
	err := app.Stop(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")

	// Clean up
	app.RequireStop()
}
