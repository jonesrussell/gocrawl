package httpd_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
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

func (m *mockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	return []any{}, nil
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	return 0, nil
}

func TestHTTPCommand(t *testing.T) {
	// Create test dependencies
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
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
			func() api.SearchManager { return mockSearch },
		),
		httpd.Module,
	)

	app.RequireStart()
	defer app.RequireStop()
}

func TestHTTPCommandGracefulShutdown(t *testing.T) {
	// Create test dependencies
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
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
			func() api.SearchManager { return mockSearch },
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
	// Create test dependencies
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

	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockSearch := &mockSearchManager{}

	// Create a test module that doesn't include api.Module
	testModule := fx.Module("test",
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() storage.Interface { return mockStore },
			func() api.SearchManager { return mockSearch },
			func() *http.Server {
				return &http.Server{
					Addr: ":-1", // Invalid port number
				}
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, server *http.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Start the server in a goroutine
					go func() {
						if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							panic(err)
						}
					}()

					// Try to listen on the port to check if it's valid
					ln, err := net.Listen("tcp", server.Addr)
					if err != nil {
						return fmt.Errorf("failed to listen on %s: %w", server.Addr, err)
					}
					ln.Close()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
		}),
	)

	// Create test app with mocked dependencies and invalid server using fxtest
	app := fxtest.New(t,
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Supply(mockSearch),
		testModule,
	)

	// Start should fail with invalid port error
	err := app.Start(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to listen on :-1: listen tcp: address -1: invalid port")

	// Cleanup
	_ = app.Stop(t.Context())
}

func TestCommand(t *testing.T) {
	cmd := httpd.Command()
	assert.NotNil(t, cmd)
	assert.Equal(t, "httpd", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func TestTLSConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		tlsConfig struct {
			Enabled     bool   `yaml:"enabled"`
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
		}
		expectError bool
	}{
		{
			name: "TLS disabled",
			tlsConfig: struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			}{
				Enabled: false,
			},
			expectError: false,
		},
		{
			name: "TLS enabled with missing certificate",
			tlsConfig: struct {
				Enabled     bool   `yaml:"enabled"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
			}{
				Enabled:     true,
				Certificate: "nonexistent.crt",
				Key:         "nonexistent.key",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := logger.NewMockInterface(ctrl)
			mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			mockCfg := &mockConfig{
				serverConfig: &config.ServerConfig{
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
						TLS: tt.tlsConfig,
					},
				},
			}

			mockStore := &mockStorage{}
			mockSearch := &mockSearchManager{}

			app := fxtest.New(t,
				fx.Supply(mockLogger),
				fx.Supply(mockCfg),
				fx.Supply(mockStore),
				fx.Supply(mockSearch),
				fx.Provide(
					func() logger.Interface { return mockLogger },
					func() config.Interface { return mockCfg },
					func() storage.Interface { return mockStore },
					func() api.SearchManager { return mockSearch },
				),
				httpd.Module,
				fx.Invoke(func(*http.Server) {}),
			)

			err := app.Start(t.Context())
			if tt.expectError {
				require.Error(t, err)
				// Check for the error in the fx error chain
				errStr := err.Error()
				assert.Contains(t, errStr, "failed to load TLS certificate")
				assert.Contains(t, errStr, "nonexistent.crt")
			} else {
				require.NoError(t, err)
				app.RequireStop()
			}
		})
	}
}

func TestServerStartStop(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create a mux for the server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	config := &config.ServerConfig{
		Address: fmt.Sprintf(":%d", port),
	}

	server := &http.Server{
		Addr:    config.Address,
		Handler: mux,
	}

	// Channel to signal when server is ready
	serverReady := make(chan struct{})
	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			serverErr <- listenErr
		}
		close(serverReady)
	}()

	// Wait for server to be ready
	select {
	case <-serverReady:
		// Server is ready, proceed with test
	case startErr := <-serverErr:
		t.Fatalf("Server failed to start: %v", startErr)
	case <-time.After(2 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Test health endpoint
	resp, respErr := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	require.NoError(t, respErr)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Stop server
	stopErr := server.Close()
	require.NoError(t, stopErr)

	// Verify server is stopped
	_, respErr = http.Get(fmt.Sprintf("http://localhost:%d/health", port))
	assert.Error(t, respErr)
}
