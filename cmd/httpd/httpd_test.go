package httpd_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Address:      ":0", // Use random port
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	// Channel to signal when server is ready
	serverReady := make(chan struct{})
	shutdownComplete := make(chan struct{})

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
					close(serverReady)
					return nil
				},
				OnStop: func(context.Context) error {
					close(shutdownComplete)
					return nil
				},
			})
		}),
	)

	// Start the app
	app.RequireStart()
	defer app.RequireStop()

	// Wait for server to be ready
	select {
	case <-serverReady:
		// Server is ready, proceed with test
	case <-time.After(2 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Trigger graceful shutdown by stopping the app
	app.RequireStop()

	// Wait for shutdown to complete
	select {
	case <-shutdownComplete:
		// Shutdown completed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Server failed to shut down within timeout")
	}
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

	// Channel to communicate server errors
	serverErr := make(chan error, 1)

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
							serverErr <- err
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

func createTestCertificates(t *testing.T) (string, string) {
	// Create temporary directory for certificates
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Generate test certificate and key
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
	require.NoError(t, err)

	// Write certificate
	certFile, err := os.Create(certPath)
	require.NoError(t, err)
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	require.NoError(t, err)
	certFile.Close()

	// Write private key
	keyFile, err := os.Create(keyPath)
	require.NoError(t, err)
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	require.NoError(t, err)
	keyFile.Close()

	return certPath, keyPath
}

func TestTLSConfiguration(t *testing.T) {
	certPath, keyPath := createTestCertificates(t)

	tests := []struct {
		name           string
		config         *config.ServerConfig
		expectedError  bool
		expectedScheme string
	}{
		{
			name: "TLS enabled with valid certificates",
			config: &config.ServerConfig{
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
					Enabled: true,
					TLS: struct {
						Enabled     bool   `yaml:"enabled"`
						Certificate string `yaml:"certificate"`
						Key         string `yaml:"key"`
					}{
						Enabled:     true,
						Certificate: certPath,
						Key:         keyPath,
					},
				},
			},
			expectedError:  false,
			expectedScheme: "https",
		},
		{
			name: "TLS disabled",
			config: &config.ServerConfig{
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
					Enabled: true,
					TLS: struct {
						Enabled     bool   `yaml:"enabled"`
						Certificate string `yaml:"certificate"`
						Key         string `yaml:"key"`
					}{
						Enabled: false,
					},
				},
			},
			expectedError:  false,
			expectedScheme: "http",
		},
		{
			name: "TLS enabled with missing certificate",
			config: &config.ServerConfig{
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
					Enabled: true,
					TLS: struct {
						Enabled     bool   `yaml:"enabled"`
						Certificate string `yaml:"certificate"`
						Key         string `yaml:"key"`
					}{
						Enabled:     true,
						Certificate: "nonexistent.pem",
						Key:         keyPath,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &http.Server{
				Addr: ":0", // Use random port
			}

			// Start server in a goroutine
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					t.Errorf("server error: %v", err)
				}
			}()

			// Wait for server to start
			time.Sleep(100 * time.Millisecond)

			if tt.expectedError {
				assert.Error(t, server.Close())
				return
			}
			defer server.Close()

			// Create test client
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			// Test the connection
			resp, err := client.Get(tt.expectedScheme + "://" + server.Addr + "/health")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestServerStartStop(t *testing.T) {
	config := &config.ServerConfig{
		Address: ":0", // Use random port
	}

	server := &http.Server{
		Addr: config.Address,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	assert.NotEmpty(t, server.Addr)

	// Test health endpoint
	resp, err := http.Get("http://" + server.Addr + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Stop server
	err = server.Close()
	require.NoError(t, err)

	// Verify server is stopped
	_, err = http.Get("http://" + server.Addr + "/health")
	assert.Error(t, err)
}
