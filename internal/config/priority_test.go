package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestConfigurationPriority(t *testing.T) {
	// Setup test environment
	cleanup := testutils.SetupTestEnv(t)
	defer cleanup()

	// Create test config file
	configContent := `
app:
  environment: test
  name: gocrawl
  version: 1.0.0
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  source_file: internal/config/testdata/sources.yml
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: config_api_key
`
	err := os.WriteFile("internal/config/testdata/config.yml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("internal/config/testdata/config.yml")

	// Create test sources file
	sourcesContent := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
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
	t.Setenv("ELASTICSEARCH_API_KEY", "env_api_key")

	// Create config
	cfg, err := config.New(testutils.NewTestLogger(t))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variable overrides config file
	esCfg := cfg.GetElasticsearchConfig()
	require.Equal(t, "env_api_key", esCfg.APIKey)
}
