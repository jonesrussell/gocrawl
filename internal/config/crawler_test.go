package config_test

import (
	"fmt"
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
			name: "valid_configuration",
			configContent: `
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
			name: "invalid_max_depth",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 0
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
			wantErr: true,
			errMsg:  "max depth must be greater than 0",
		},
		{
			name: "invalid_parallelism",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
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
			errMsg:  "parallelism must be greater than 0",
		},
		{
			name: "invalid_rate_limit",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: invalid
  parallelism: 2
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: true,
			errMsg:  "invalid rate limit duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := testutils.SetupTestEnvironment(t, tt.configContent, tt.sourcesContent)
			defer setup.Cleanup()

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Validate specific fields
			crawlerCfg := cfg.GetCrawlerConfig()
			require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
			require.Equal(t, 2, crawlerCfg.MaxDepth)
			require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
			require.Equal(t, 2, crawlerCfg.Parallelism)
			require.Equal(t, setup.SourcesPath, crawlerCfg.SourceFile)
		})
	}
}

func TestCrawlerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) *testutils.TestSetup
		wantErr bool
	}{
		{
			name: "invalid max depth",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
crawler:
  max_depth: 0
`, "")
			},
			wantErr: true,
		},
		{
			name: "invalid parallelism",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
crawler:
  parallelism: 0
`, "")
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
crawler:
  rate_limit: 0s
`, "")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := tt.setup(t)
			defer setup.Cleanup()

			cfg, err := config.LoadConfig(setup.ConfigPath)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
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

func TestSetMaxDepth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		depth   int
		wantErr bool
	}{
		{
			name:    "valid depth",
			depth:   2,
			wantErr: false,
		},
		{
			name:    "zero depth",
			depth:   0,
			wantErr: true,
		},
		{
			name:    "negative depth",
			depth:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setup := testutils.SetupTestEnvironment(t, `
crawler:
  max_depth: `+fmt.Sprint(tt.depth)+`
  base_url: http://test.example.com
`, "")
			defer setup.Cleanup()

			cfg, err := config.LoadConfig(setup.ConfigPath)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.depth, cfg.Crawler.MaxDepth)
		})
	}
}

func TestSetRateLimit(t *testing.T) {
	setup := testutils.SetupTestEnvironment(t, "", "")
	defer setup.Cleanup()

	cfg := &config.CrawlerConfig{}
	rateLimit := 2 * time.Second
	cfg.SetRateLimit(rateLimit)
	require.Equal(t, rateLimit, cfg.RateLimit)
	require.Equal(t, "2s", viper.GetString("crawler.rate_limit"))
}

func TestSetBaseURL(t *testing.T) {
	setup := testutils.SetupTestEnvironment(t, "", "")
	defer setup.Cleanup()

	cfg := &config.CrawlerConfig{}
	url := "http://example.com"
	cfg.SetBaseURL(url)
	require.Equal(t, url, cfg.BaseURL)
	require.Equal(t, url, viper.GetString("crawler.base_url"))
}

func TestSetIndexName(t *testing.T) {
	setup := testutils.SetupTestEnvironment(t, "", "")
	defer setup.Cleanup()

	cfg := &config.CrawlerConfig{}
	index := "test_index"
	cfg.SetIndexName(index)
	require.Equal(t, index, cfg.IndexName)
	require.Equal(t, index, viper.GetString("elasticsearch.index_name"))
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
			setup := testutils.SetupTestEnvironment(t, "", "")
			defer setup.Cleanup()

			viper.Reset()
			cfg := &config.CrawlerConfig{}
			tt.setup(cfg)
			tt.validate(t, cfg)
		})
	}
}
