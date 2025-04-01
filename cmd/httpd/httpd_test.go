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
	commontypes "github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
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
		return nil, nil
	}
	return val, nil
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

func (w *searchManagerWrapper) Close() error {
	return nil
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
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
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
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
	defer app.RequireStop()
}

func TestHTTPCommandGracefulShutdown(t *testing.T) {
	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
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
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
	defer app.RequireStop()
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
	mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
	mockCfg.On("GetSources").Return([]config.Source{}, nil)
	mockCfg.On("GetCommand").Return("test")

	mockStore := &mockStorage{}
	mockSecurity := &testutils.MockSecurityMiddleware{}

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
	)

	app.RequireStart()
	defer app.RequireStop()

	// Test health check endpoint
	client := &http.Client{
		Timeout: api.HealthCheckInterval,
	}
	resp, err := client.Get(fmt.Sprintf("http://%s/health", mockCfg.GetServerConfig().Address))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
	)

	app.RequireStart()
	defer app.RequireStop()

	// Verify storage connection was tested
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
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
	)

	// Verify that the app fails to start due to storage connection error
	err := app.Start(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to storage")
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
				func() commontypes.Logger { return mockLogger },
				fx.As(new(commontypes.Logger)),
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
