package httpd_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockLogger implements logger.Interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(_ string, _ ...any)       {}
func (m *mockLogger) Info(_ string, _ ...any)        {}
func (m *mockLogger) Warn(_ string, _ ...any)        {}
func (m *mockLogger) Error(_ string, _ ...any)       {}
func (m *mockLogger) Fatal(_ string, _ ...any)       {}
func (m *mockLogger) Panic(_ string, _ ...any)       {}
func (m *mockLogger) With(_ ...any) logger.Interface { return m }
func (m *mockLogger) Errorf(_ string, _ ...any)      {}
func (m *mockLogger) Printf(_ string, _ ...any)      {}
func (m *mockLogger) Sync() error                    { return nil }

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Address:      ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func TestModule(t *testing.T) {
	// Create mock dependencies
	mockLogger := &mockLogger{}

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
		),
		httpd.Module,
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
	require.NoError(t, app.Err())
}

// TestModuleProvides tests that the httpd module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	var server *http.Server

	app := fxtest.New(t,
		httpd.Module,
		fx.Provide(
			func() logger.Interface { return &mockLogger{} },
			func() api.SearchManager { return &mockSearchManager{} },
			func() config.Interface { return &mockConfig{} },
		),
		fx.Populate(&server),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, server)
}

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}
