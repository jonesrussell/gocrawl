package config_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) *testutils.TestSetup
		validate func(*testing.T, config.Interface, error)
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
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
log:
  level: debug
elasticsearch:
  addresses:
    - https://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  tls:
    enabled: true
    certificate: test-cert.pem
    key: test-key.pem
`, "")
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

				esCfg := cfg.GetElasticsearchConfig()
				require.Equal(t, []string{"https://localhost:9200"}, esCfg.Addresses)
				require.Equal(t, "id:test_api_key", esCfg.APIKey)
				require.Equal(t, "test-index", esCfg.IndexName)
				require.True(t, esCfg.TLS.Enabled)
				require.Equal(t, "test-cert.pem", esCfg.TLS.CertFile)
				require.Equal(t, "test-key.pem", esCfg.TLS.KeyFile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			setup := tt.setup(t)
			defer setup.Cleanup()

			// Configure Viper
			viper.SetConfigFile(setup.ConfigPath)
			viper.SetConfigType("yaml")
			err := viper.ReadInConfig()
			require.NoError(t, err)

			// Create config
			cfg, err := config.New(testutils.NewTestLogger(t))

			// Validate results
			tt.validate(t, cfg, err)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedError string
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Environment: "development",
				App: &app.Config{
					Environment: "development",
					Name:        "test",
					Version:     "1.0.0",
				},
				Logger:   log.New(),
				Server:   server.New(),
				Priority: priority.New(),
				Crawler:  config.NewCrawlerConfig(),
				Sources: []config.Source{
					{
						Name:           "test",
						URL:            "http://test.example.com",
						AllowedDomains: []string{"test.example.com"},
						StartURLs:      []string{"http://test.example.com"},
						RateLimit:      time.Second,
						MaxDepth:       1,
						ArticleIndex:   "articles",
						Index:          "test",
						Selectors: config.SourceSelectors{
							Article: config.ArticleSelectors{
								Title: "h1",
								Body:  "article",
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "missing allowed domains",
			config: &config.Config{
				Environment: "development",
				App: &app.Config{
					Environment: "development",
					Name:        "test",
					Version:     "1.0.0",
				},
				Logger:   log.New(),
				Server:   server.New(),
				Priority: priority.New(),
				Crawler:  config.NewCrawlerConfig(),
				Sources: []config.Source{
					{
						Name:         "test",
						URL:          "http://test.example.com",
						StartURLs:    []string{"http://test.example.com"},
						RateLimit:    time.Second,
						MaxDepth:     1,
						ArticleIndex: "articles",
						Index:        "test",
						Selectors: config.SourceSelectors{
							Article: config.ArticleSelectors{
								Title: "h1",
								Body:  "article",
							},
						},
					},
				},
			},
			expectedError: "missing allowed domains",
		},
		{
			name: "missing start URLs",
			config: &config.Config{
				Environment: "development",
				App: &app.Config{
					Environment: "development",
					Name:        "test",
					Version:     "1.0.0",
				},
				Logger:   log.New(),
				Server:   server.New(),
				Priority: priority.New(),
				Crawler:  config.NewCrawlerConfig(),
				Sources: []config.Source{
					{
						Name:           "test",
						URL:            "http://test.example.com",
						AllowedDomains: []string{"test.example.com"},
						RateLimit:      time.Second,
						MaxDepth:       1,
						ArticleIndex:   "articles",
						Index:          "test",
						Selectors: config.SourceSelectors{
							Article: config.ArticleSelectors{
								Title: "h1",
								Body:  "article",
							},
						},
					},
				},
			},
			expectedError: "missing start URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSource_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  config.Source
		wantErr bool
	}{
		{
			name: "valid source",
			source: config.Source{
				AllowedDomains: []string{"example.com"},
				StartURLs:      []string{"https://example.com"},
				Rules: config.Rules{
					{
						Pattern:  "/blog/*",
						Action:   config.ActionAllow,
						Priority: 1,
					},
					{
						Pattern:  "/admin/*",
						Action:   config.ActionDisallow,
						Priority: 2,
					},
				},
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{
						Title:   "h1",
						Body:    "article",
						TimeAgo: "time",
						Author:  ".author",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing allowed domains",
			source: config.Source{
				StartURLs: []string{"https://example.com"},
				Rules: config.Rules{
					{
						Pattern:  "/blog/*",
						Action:   config.ActionAllow,
						Priority: 1,
					},
				},
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{
						Title: "h1",
						Body:  "article",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing start URLs",
			source: config.Source{
				AllowedDomains: []string{"example.com"},
				Rules: config.Rules{
					{
						Pattern:  "/blog/*",
						Action:   config.ActionAllow,
						Priority: 1,
					},
				},
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{
						Title: "h1",
						Body:  "article",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid URL in start URLs",
			source: config.Source{
				AllowedDomains: []string{"example.com"},
				StartURLs:      []string{"not-a-url"},
				Rules: config.Rules{
					{
						Pattern:  "/blog/*",
						Action:   config.ActionAllow,
						Priority: 1,
					},
				},
				Selectors: config.SourceSelectors{
					Article: config.ArticleSelectors{
						Title: "h1",
						Body:  "article",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.source.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidateSelectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		selectors config.ArticleSelectors
		wantErr   bool
	}{
		{
			name: "valid selectors",
			selectors: config.ArticleSelectors{
				Title:   "h1",
				Body:    "article",
				TimeAgo: "time",
				Author:  ".author",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			selectors: config.ArticleSelectors{
				Body:    "article",
				TimeAgo: "time",
				Author:  ".author",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			selectors: config.ArticleSelectors{
				Title:   "h1",
				TimeAgo: "time",
				Author:  ".author",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			selectors: config.ArticleSelectors{
				Title:   "",
				Body:    "article",
				TimeAgo: "time",
				Author:  ".author",
			},
			wantErr: true,
		},
		{
			name: "empty body",
			selectors: config.ArticleSelectors{
				Title:   "h1",
				Body:    "",
				TimeAgo: "time",
				Author:  ".author",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.selectors.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRules_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rules   config.Rules
		wantErr bool
	}{
		{
			name: "valid rules",
			rules: config.Rules{
				{
					Pattern:  "/blog/*",
					Action:   config.ActionAllow,
					Priority: 1,
				},
				{
					Pattern:  "/admin/*",
					Action:   config.ActionDisallow,
					Priority: 2,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty rules",
			rules:   config.Rules{},
			wantErr: false,
		},
		{
			name: "invalid pattern",
			rules: config.Rules{
				{
					Pattern:  "",
					Action:   config.ActionAllow,
					Priority: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid action",
			rules: config.Rules{
				{
					Pattern:  "/blog/*",
					Action:   "invalid",
					Priority: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "negative priority",
			rules: config.Rules{
				{
					Pattern:  "/blog/*",
					Action:   config.ActionAllow,
					Priority: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.rules.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
