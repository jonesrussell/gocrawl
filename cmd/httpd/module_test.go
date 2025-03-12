package httpd_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockLogger := logger.NewMockInterface(ctrl)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
		),
		httpd.Module,
	)
	require.NoError(t, app.Err())
}

// TestModuleProvides tests that the httpd module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	var server *http.Server

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockSearch := &mockSearchManager{}
	mockCfg := &mockConfig{}

	app := fxtest.New(t,
		httpd.Module,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.SearchManager { return mockSearch },
			func() config.Interface { return mockCfg },
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
