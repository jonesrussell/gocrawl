package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestConfigurationPriority(t *testing.T) {
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
  environment: development
  name: gocrawl
  version: 1.0.0
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: test_api_key
crawler:
  source_file: ` + sourcesPath + `
  base_url: http://test.example.com
  max_depth: 1
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("CONFIG_FILE", configPath)
	t.Setenv("APP_ENVIRONMENT", "production")
	t.Setenv("APP_NAME", "gocrawl-env")
	t.Setenv("APP_VERSION", "2.0.0")

	// Create config
	cfg, err := config.New(testutils.NewTestLogger(t))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variable overrides config file
	esCfg := cfg.GetElasticsearchConfig()
	require.Equal(t, "test_api_key", esCfg.APIKey)
}
