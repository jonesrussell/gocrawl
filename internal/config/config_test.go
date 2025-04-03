package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

// testLogger implements config.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l testLogger) Info(msg string, fields ...config.Field) {
	l.t.Logf("INFO: %s %v", msg, fields)
}

func (l testLogger) Warn(msg string, fields ...config.Field) {
	l.t.Logf("WARN: %s %v", msg, fields)
}

// newTestLogger creates a new test logger
func newTestLogger(t *testing.T) config.Logger {
	return testLogger{t: t}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Create temporary test directory
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				sourcesPath := filepath.Join(tmpDir, "sources.yml")

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
sources:
  - name: test
    url: http://test.example.com
    rate_limit: 100ms
    max_depth: 1
`
				err := os.WriteFile(configPath, []byte(configContent), 0644)
				require.NoError(t, err)

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
				err = os.WriteFile(sourcesPath, []byte(sourcesContent), 0644)
				require.NoError(t, err)

				// Set environment variables
				t.Setenv("CONFIG_FILE", configPath)
			},
			validate: func(t *testing.T, cfg config.Interface, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)

				// Verify configuration
				appCfg := cfg.GetAppConfig()
				require.Equal(t, "test", appCfg.Environment)
				require.Equal(t, "gocrawl", appCfg.Name)
				require.Equal(t, "1.0.0", appCfg.Version)
				require.False(t, appCfg.Debug)

				crawlerCfg := cfg.GetCrawlerConfig()
				require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
				require.Equal(t, 2, crawlerCfg.MaxDepth)
				require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
				require.Equal(t, 2, crawlerCfg.Parallelism)

				logCfg := cfg.GetLogConfig()
				require.Equal(t, "debug", logCfg.Level)
				require.True(t, logCfg.Debug)

				esCfg := cfg.GetElasticsearchConfig()
				require.Equal(t, []string{"https://localhost:9200"}, esCfg.Addresses)
				require.Equal(t, "test_api_key", esCfg.APIKey)
				require.True(t, esCfg.TLS.Enabled)
				require.Equal(t, "test-cert.pem", esCfg.TLS.CertFile)
				require.Equal(t, "test-key.pem", esCfg.TLS.KeyFile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := testutils.SetupTestEnv(t)
			defer cleanup()

			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.New(testutils.NewTestLogger(t))

			// Validate results
			tt.validate(t, cfg, err)
		})
	}
}
