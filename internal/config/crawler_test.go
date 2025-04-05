package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestCrawlerConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		sourcesContent string
		wantErr        bool
		errMsg         string
	}{
		{
			name: "valid configuration",
			configContent: `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
  debug: false
log:
  level: info
  debug: false
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: %s
  random_delay: 1s
  index_name: gocrawl
  content_index_name: gocrawl-content
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
  api_key: id:key
server:
  address: :8080
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 15s
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    selectors:
      article:
        container: article
        title: h1
        body: .content`,
			wantErr: false,
		},
		{
			name: "invalid configuration",
			configContent: `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
  debug: false
log:
  level: info
  debug: false
crawler:
  base_url: ""
  max_depth: 0
  rate_limit: invalid
  parallelism: 0
  source_file: %s
  random_delay: 1s
  index_name: gocrawl
  content_index_name: gocrawl-content
elasticsearch:
  addresses:
    - http://localhost:9200
  index_name: gocrawl
  api_key: id:key
server:
  address: :8080
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 15s
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    selectors:
      article:
        container: article
        title: h1
        body: .content`,
			wantErr: true,
			errMsg:  "crawler base URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test files
			tmpDir, err := os.MkdirTemp("", "gocrawl-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Create sources file path
			sourcesPath := filepath.Join(tmpDir, "sources.yml")

			// Create test setup
			setup := testutils.SetupTestEnvironment(t,
				fmt.Sprintf(tt.configContent, sourcesPath),
				tt.sourcesContent)
			defer setup.Cleanup()

			t.Logf("Config file path: %s", setup.ConfigPath)
			t.Logf("Sources file path: %s", setup.SourcesPath)
			t.Logf("Config content: %s", tt.configContent)

			// Load and validate config
			cfg, err := config.LoadConfig(setup.ConfigPath)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			t.Logf("Loaded config: %+v", cfg)
			t.Logf("Crawler config: %+v", cfg.GetCrawlerConfig())

			// Verify crawler config
			crawlerCfg := cfg.GetCrawlerConfig()
			require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
			require.Equal(t, 2, crawlerCfg.MaxDepth)
			require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
			require.Equal(t, 2, crawlerCfg.Parallelism)
			require.Equal(t, setup.SourcesPath, crawlerCfg.SourceFile)
			require.Equal(t, time.Second, crawlerCfg.RandomDelay)
			require.Equal(t, "gocrawl", crawlerCfg.IndexName)
			require.Equal(t, "gocrawl-content", crawlerCfg.ContentIndexName)
		})
	}
}

func TestParseRateLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid duration",
			input:   "2s",
			want:    2 * time.Second,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
			errMsg:  "rate limit cannot be empty",
		},
		{
			name:    "invalid duration",
			input:   "invalid",
			want:    0,
			wantErr: true,
			errMsg:  "error parsing duration",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := config.ParseRateLimit(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCrawlerConfig_Setters(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*config.CrawlerConfig)
		validate func(*testing.T, *config.CrawlerConfig)
	}{
		{
			name: "set max depth",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetMaxDepth(5)
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, 5, cfg.MaxDepth)
			},
		},
		{
			name: "set rate limit",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetRateLimit(2 * time.Second)
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, 2*time.Second, cfg.RateLimit)
			},
		},
		{
			name: "set base url",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetBaseURL("http://example.com")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "http://example.com", cfg.BaseURL)
			},
		},
		{
			name: "set index name",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetIndexName("test_index")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "test_index", cfg.IndexName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := testutils.SetupTestEnvironment(t, fmt.Sprintf(`
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
  debug: false
crawler:
  source_file: %s
`, "sources.yml"), "")
			defer setup.Cleanup()

			cfg := &config.CrawlerConfig{}
			tt.setup(cfg)
			tt.validate(t, cfg)
		})
	}
}
