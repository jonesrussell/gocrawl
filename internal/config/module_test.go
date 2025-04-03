package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// TestModule provides tests for the config module's dependency injection.
func TestModule(t *testing.T) {
	// Create test config file
	configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: internal/config/testdata/sources.yml
logging:
  level: debug
  debug: true
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
`
	err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("internal/config/testdata/config.yml")

	// Create test sources file
	sourcesContent := `
sources:
  test_source:
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    article_index: test_articles
    content_index: test_content
    selectors:
      title: h1
      content: article
      author: .author
      date: .date
`
	err = os.WriteFile("internal/config/testdata/sources.yml", []byte(sourcesContent), 0644)
	require.NoError(t, err)
	defer os.Remove("internal/config/testdata/sources.yml")

	// Set environment variables
	t.Setenv("CONFIG_FILE", "internal/config/testdata/config.yml")

	// Create test application
	app := fxtest.New(t,
		fx.Provide(
			func() config.Interface {
				cfg, err := config.New(newTestLogger(t))
				require.NoError(t, err)
				return cfg
			},
			newTestLogger,
		),
		fx.Invoke(func(cfg config.Interface) {
			require.NotNil(t, cfg)
			appCfg := cfg.GetAppConfig()
			require.Equal(t, "test", appCfg.Environment)
			require.Equal(t, "gocrawl", appCfg.Name)
			require.Equal(t, "1.0.0", appCfg.Version)
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

// TestModuleLifecycle tests the module's lifecycle hooks.
func TestModuleLifecycle(t *testing.T) {
	app := fxtest.New(t,
		fx.Provide(
			func() config.Interface {
				return config.NewNoOp()
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, cfg config.Interface) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					require.NotNil(t, cfg)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	app.RequireStart()
	app.RequireStop()
}
