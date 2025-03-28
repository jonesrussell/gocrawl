// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
)

// TestCommonModule provides a test-specific common module that excludes the logger module.
var TestCommonModule = fx.Module("testCommon",
	// Suppress fx logging to reduce noise in the application logs.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),
	// Core modules used by most commands, excluding logger and config.
	sources.Module,
)

// TestConfigModule provides a test-specific config module that doesn't try to load files.
var TestConfigModule = fx.Module("testConfig",
	fx.Provide(
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
				mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				return mockCfg
			},
			fx.ResultTags(`name:"config"`),
		),
	),
)

// setupTestApp creates a new test application with all required dependencies.
// It provides mock implementations of all interfaces required by the crawler module.
func setupTestApp(t *testing.T) *fxtest.App {
	t.Helper()

	// Set environment variables to prevent file loading
	t.Setenv("APP_ENV", "test")
	t.Setenv("CONFIG_PATH", "")

	// Create mock logger
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	return fxtest.New(t,
		fx.NopLogger,
		// Provide core dependencies
		fx.Provide(
			// Named dependencies
			func() sources.Interface { return &mockSources{} },
			// Logger provider that replaces the default logger.Module provider
			fx.Annotate(
				func() common.Logger { return mockLogger },
				fx.ResultTags(`name:"logger"`),
			),
			// Provide unnamed interfaces required by the crawler module
			func() storagetypes.Interface { return &mockStorage{} },
			fx.Annotate(
				func() api.IndexManager { return &mockIndexManager{} },
				fx.ResultTags(`name:"indexManager"`),
			),
			func() api.SearchManager { return &mockSearchManager{} },
		),
		// Mock content processors
		fx.Provide(
			fx.Annotate(
				func() collector.Processor {
					return &mockContentProcessor{}
				},
				fx.ResultTags(`group:"processors"`),
			),
		),
		// Include test config module and crawler module
		TestConfigModule,
		TestCommonModule,
		crawler.Module,
		// Verify dependencies
		fx.Invoke(func(p crawler.Params) {
			verifyDependencies(t, &p)
		}),
	)
}

// verifyDependencies checks that all required dependencies are present in the Params struct.
func verifyDependencies(t *testing.T, p *crawler.Params) {
	t.Helper()

	require.NotNil(t, p)
	require.NotNil(t, p.Logger)
	require.NotNil(t, p.Sources)
	require.NotNil(t, p.IndexManager)
}

// TestDependencyInjection verifies that all dependencies are properly injected into the Params struct.
func TestDependencyInjection(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestModuleConstruction verifies that the crawler module can be constructed with all required dependencies.
func TestModuleConstruction(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestModuleLifecycle verifies that the crawler module can be started and stopped correctly.
func TestModuleLifecycle(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	require.NoError(t, app.Stop(t.Context()))
}
