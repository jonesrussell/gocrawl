package httpd_test

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// mockStorage implements storage.Interface for testing
type mockStorage struct {
	storage.Interface
}

func (m *mockStorage) Search(ctx context.Context, query string, opts any) ([]any, error) {
	return []any{}, nil
}

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Address:      ":0", // Use random port
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func TestHTTPCommand(t *testing.T) {
	// Create test dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}

	// Create test app with mocked dependencies
	app := fx.New(
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
		),
		httpd.Module,
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	require.NoError(t, app.Stop(t.Context()))
}

func TestHTTPCommandGracefulShutdown(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Create test dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	// Set up logger expectations
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}

	// Channel to signal when server is ready
	serverReady := make(chan struct{})

	// Create test app with mocked dependencies
	app := fx.New(
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
		),
		httpd.Module,
		fx.Invoke(func(server *http.Server) {
			// Signal that server is ready to accept connections
			go func() {
				// Wait for server to be assigned a port
				time.Sleep(100 * time.Millisecond)
				close(serverReady)
			}()
		}),
	)

	// Start the app
	require.NoError(t, app.Start(ctx))

	// Wait for server to be ready
	select {
	case <-serverReady:
		// Server is ready, proceed with test
	case <-time.After(2 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Simulate SIGTERM
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGTERM))

	// Give the server time to shut down gracefully
	time.Sleep(500 * time.Millisecond)

	// Stop the app
	require.NoError(t, app.Stop(t.Context()))
}

func TestHTTPCommandServerError(t *testing.T) {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Create test dependencies with an invalid port to force an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}

	// Create test app with mocked dependencies and invalid server
	app := fx.New(
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
			func() *http.Server {
				return &http.Server{
					Addr: ":-1", // Invalid port number
				}
			},
		),
		httpd.Module,
	)

	// Start the app - should fail due to invalid port
	err := app.Start(ctx)
	require.Error(t, err)

	// Stop the app
	require.NoError(t, app.Stop(t.Context()))
}

func TestCommand(t *testing.T) {
	cmd := httpd.Command()
	assert.NotNil(t, cmd)
	assert.Equal(t, "httpd", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}
