package config_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule(t *testing.T) {
	t.Parallel()

	// Create test app with config module
	app := fxtest.New(t,
		fx.Supply(config.Params{
			Environment: "test",
			Debug:       true,
			Command:     "test",
		}),
		config.Module,
		fx.Invoke(func(c config.Interface) {
			// Test that we can get a config instance
			assert.NotNil(t, c)

			// Test app config
			appConfig := c.GetAppConfig()
			assert.Equal(t, "test", appConfig.Environment)
			assert.Equal(t, "gocrawl", appConfig.Name)
			assert.Equal(t, "1.0.0", appConfig.Version)
			assert.False(t, appConfig.Debug)

			// Test log config
			logConfig := c.GetLogConfig()
			assert.Equal(t, "info", logConfig.Level)
			assert.False(t, logConfig.Debug)

			// Test elasticsearch config
			esConfig := c.GetElasticsearchConfig()
			assert.Equal(t, []string{"http://localhost:9200"}, esConfig.Addresses)
			assert.Equal(t, "gocrawl", esConfig.IndexName)

			// Test server config
			serverConfig := c.GetServerConfig()
			assert.Equal(t, ":8080", serverConfig.Address)

			// Test sources
			sources, err := c.GetSources()
			require.NoError(t, err)
			assert.Empty(t, sources)

			// Test command
			assert.Equal(t, "test", c.GetCommand())
		}),
	)

	// Start the app
	app.RequireStart()
	app.RequireStop()
}

func TestNewNoOp(t *testing.T) {
	t.Parallel()

	// Create a no-op config
	c := config.NewNoOp()

	// Test app config
	appConfig := c.GetAppConfig()
	assert.Equal(t, "test", appConfig.Environment)
	assert.Equal(t, "gocrawl", appConfig.Name)
	assert.Equal(t, "1.0.0", appConfig.Version)
	assert.False(t, appConfig.Debug)

	// Test log config
	logConfig := c.GetLogConfig()
	assert.Equal(t, "info", logConfig.Level)
	assert.False(t, logConfig.Debug)

	// Test elasticsearch config
	esConfig := c.GetElasticsearchConfig()
	assert.Equal(t, []string{"http://localhost:9200"}, esConfig.Addresses)
	assert.Equal(t, "gocrawl", esConfig.IndexName)

	// Test server config
	serverConfig := c.GetServerConfig()
	assert.Equal(t, ":8080", serverConfig.Address)

	// Test sources
	sources, err := c.GetSources()
	require.NoError(t, err)
	assert.Empty(t, sources)

	// Test command
	assert.Equal(t, "test", c.GetCommand())
}
