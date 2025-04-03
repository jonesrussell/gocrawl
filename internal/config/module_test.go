package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// TestModule provides tests for the config module's dependency injection.
func TestModule(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	sourcesPath := filepath.Join(tmpDir, "sources.yml")

	// Create test sources file
	sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
	err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
	require.NoError(t, err)

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
  source_file: ` + sourcesPath + `
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
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("CONFIG_FILE", configPath)

	// Create test application
	app := fxtest.New(t,
		fx.Provide(
			newTestLogger,
			config.New,
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
	// Create temporary test directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	sourcesPath := filepath.Join(tmpDir, "sources.yml")

	// Create test sources file
	sourcesContent := `
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
    selectors:
      article:
        title: h1
        body: article
`
	err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
	require.NoError(t, err)

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
  source_file: ` + sourcesPath + `
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
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("CONFIG_FILE", configPath)

	app := fxtest.New(t,
		fx.Provide(
			newTestLogger,
			config.New,
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
