package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

// TestModule provides tests for the config module's dependency injection.
func TestModule(t *testing.T) {
	// Get the absolute path to the testdata directory
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Set up test environment
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Create test application
	app := fxtest.New(t,
		fx.Provide(
			fx.Annotate(
				func() *testing.T { return t },
				fx.ResultTags(`name:"test"`),
			),
			fx.Annotate(
				testutils.NewTestLogger,
				fx.ParamTags(`name:"test"`),
			),
			config.New,
		),
		fx.Invoke(func(cfg config.Interface) {
			require.NotNil(t, cfg)
			appCfg := cfg.GetAppConfig()
			require.Equal(t, "development", appCfg.Environment)
			require.Equal(t, "gocrawl-test", appCfg.Name)
			require.Equal(t, "0.0.1", appCfg.Version)
		}),
	)

	app.RequireStart()
	app.RequireStop()
}

// TestNewNoOp tests the no-op config implementation.
func TestNewNoOp(t *testing.T) {
	t.Parallel()

	// Create a no-op config
	c := config.NewNoOp()

	// Test app config
	appConfig := c.GetAppConfig()
	require.Equal(t, "test", appConfig.Environment)
	require.Equal(t, "gocrawl", appConfig.Name)
	require.Equal(t, "1.0.0", appConfig.Version)
	require.False(t, appConfig.Debug)

	// Test log config
	logConfig := c.GetLogConfig()
	require.Equal(t, "info", logConfig.Level)
	require.False(t, logConfig.Debug)

	// Test elasticsearch config
	esConfig := c.GetElasticsearchConfig()
	require.Equal(t, []string{"http://localhost:9200"}, esConfig.Addresses)
	require.Equal(t, "gocrawl", esConfig.IndexName)

	// Test server config
	serverConfig := c.GetServerConfig()
	require.Equal(t, ":8080", serverConfig.Address)

	// Test sources
	sources := c.GetSources()
	require.Empty(t, sources)

	// Test command
	require.Equal(t, "test", c.GetCommand())
}

// TestModuleLifecycle tests the lifecycle of the config module.
func TestModuleLifecycle(t *testing.T) {
	// Get the absolute path to the testdata directory
	testdataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	configPath := filepath.Join(testdataDir, "config.yml")
	sourcesPath := filepath.Join(testdataDir, "sources.yml")

	// Verify test files exist
	require.FileExists(t, configPath, "config.yml should exist in testdata directory")
	require.FileExists(t, sourcesPath, "sources.yml should exist in testdata directory")

	// Set up test environment
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Create test module
	module := fx.Module("test",
		fx.Provide(
			fx.Annotate(
				func() *testing.T { return t },
				fx.ResultTags(`name:"test"`),
			),
			fx.Annotate(
				testutils.NewTestLogger,
				fx.ParamTags(`name:"test"`),
			),
			config.New,
		),
		fx.Invoke(func(cfg config.Interface) {
			require.NotNil(t, cfg)
		}),
	)

	// Create test app
	app := fxtest.New(t,
		module,
	)

	// Start app
	app.RequireStart()

	// Stop app
	app.RequireStop()
}
