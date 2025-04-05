package config_test

import (
	"testing"
	"time"

	"github.com/spf13/viper"
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
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: false,
		},
		{
			name: "invalid configuration",
			configContent: `
app:
  environment: test
crawler:
  base_url: http://test.example.com
  max_depth: 0
  rate_limit: invalid
  parallelism: 0
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: true,
			errMsg:  "max depth must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := testutils.SetupTestEnvironment(t, tt.configContent, tt.sourcesContent)
			defer setup.Cleanup()

			viper.Reset()
			viper.Set("crawler.source_file", setup.SourcesPath)
			viper.Set("app.environment", "test")

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			crawlerCfg := cfg.GetCrawlerConfig()
			require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
			require.Equal(t, 2, crawlerCfg.MaxDepth)
			require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
			require.Equal(t, 2, crawlerCfg.Parallelism)
			require.Equal(t, setup.SourcesPath, crawlerCfg.SourceFile)
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
			want:    config.DefaultRateLimit,
			wantErr: false,
		},
		{
			name:    "invalid duration",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := config.ParseRateLimit(tt.input)
			if tt.wantErr {
				require.Error(t, err)
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
				require.Equal(t, 5, viper.GetInt("crawler.max_depth"))
			},
		},
		{
			name: "set rate limit",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetRateLimit(2 * time.Second)
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, 2*time.Second, cfg.RateLimit)
				require.Equal(t, "2s", viper.GetString("crawler.rate_limit"))
			},
		},
		{
			name: "set base url",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetBaseURL("http://example.com")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "http://example.com", cfg.BaseURL)
				require.Equal(t, "http://example.com", viper.GetString("crawler.base_url"))
			},
		},
		{
			name: "set index name",
			setup: func(cfg *config.CrawlerConfig) {
				cfg.SetIndexName("test_index")
			},
			validate: func(t *testing.T, cfg *config.CrawlerConfig) {
				require.Equal(t, "test_index", cfg.IndexName)
				require.Equal(t, "test_index", viper.GetString("elasticsearch.index_name"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := testutils.SetupTestEnvironment(t, `
app:
  environment: test
crawler:
  source_file: sources.yml
`, "")
			defer setup.Cleanup()

			viper.Reset()
			viper.Set("crawler.source_file", setup.SourcesPath)
			viper.Set("app.environment", "test")

			cfg := &config.CrawlerConfig{}
			tt.setup(cfg)
			tt.validate(t, cfg)
		})
	}
}
