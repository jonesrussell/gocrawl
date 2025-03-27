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
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockStorage implements types.Interface for testing
type mockStorage struct {
	types.Interface
}

func (m *mockStorage) Search(context.Context, string, any) ([]any, error) {
	return []any{}, nil
}

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
	serverConfig *config.ServerConfig
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	if m.serverConfig != nil {
		return m.serverConfig
	}
	return &config.ServerConfig{
		Address:      ":0",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
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
			TLS: struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			}{
				Enabled: false,
			},
		},
	}
}

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}

func (m *mockSearchManager) Search(context.Context, string, any) ([]any, error) {
	return []any{}, nil
}

func (m *mockSearchManager) Count(context.Context, string, any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Close() error {
	return nil
}

func TestHTTPCommand(t *testing.T) {
	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockSearch := &mockSearchManager{}

	// Create test app with mocked dependencies using fxtest
	app := fxtest.New(t,
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Supply(mockSearch),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() types.Interface { return mockStore },
			fx.Annotate(
				func() api.SearchManager { return mockSearch },
				fx.ResultTags(`name:"searchManager"`),
			),
		),
		httpd.Module,
	)

	app.RequireStart()
	defer app.RequireStop()
}

func TestHTTPCommandGracefulShutdown(t *testing.T) {
	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	// Set up logger expectations for both single and multi-argument calls
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockSearch := &mockSearchManager{}

	// Create test app with mocked dependencies using fxtest
	app := fxtest.New(t,
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Supply(mockSearch),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() types.Interface { return mockStore },
			fx.Annotate(
				func() api.SearchManager { return mockSearch },
				fx.ResultTags(`name:"searchManager"`),
			),
		),
		httpd.Module,
		fx.Invoke(func(lc fx.Lifecycle, server *http.Server) {
			// Signal that server is ready to accept connections
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app
	app.RequireStart()
	defer app.RequireStop()
}

func TestHTTPCommandServerError(t *testing.T) {
	app := fxtest.New(t,
		fx.Supply(
			&testutils.MockLogger{},
			testutils.NewMockConfig(),
			&testutils.MockSecurityMiddleware{},
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			fx.Annotate(
				func() api.SearchManager { return &testutils.MockSearchManager{} },
				fx.As(new(api.SearchManager)),
			),
		),
		api.Module,
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
