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
	storagetypes.Interface
}

func (m *mockStorage) Search(context.Context, string, any) ([]any, error) {
	return []any{}, nil
}

func (m *mockStorage) Close() error {
	return nil
}

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
		fx.Supply(
			fx.Annotate(mockLogger, fx.As(new(commontypes.Logger))),
			mockStore,
			mockSecurity,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
		),
		httpd.Module,
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
		fx.Supply(
			fx.Annotate(mockLogger, fx.As(new(commontypes.Logger))),
			mockStore,
			mockSecurity,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
		),
		httpd.Module,
	)

	app.RequireStart()
	defer app.RequireStop()
}

func TestHTTPCommandServerError(t *testing.T) {
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
		fx.Supply(
			fx.Annotate(mockLogger, fx.As(new(commontypes.Logger))),
			mockStore,
			mockSecurity,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
		),
		httpd.Module,
	)

	err := app.Start(t.Context())
	require.Error(t, err)
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
