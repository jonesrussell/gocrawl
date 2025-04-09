package config_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := config.New()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.GetAppConfig())
	require.NotNil(t, cfg.GetLogConfig())
	require.NotNil(t, cfg.GetElasticsearchConfig())
	require.NotNil(t, cfg.GetServerConfig())
	require.NotNil(t, cfg.GetPriorityConfig())
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
				Logger: &log.Config{
					Level: "debug",
				},
				Server: &server.Config{
					Address: ":8080",
				},
				Priority: &priority.Config{
					DefaultPriority: 1,
					Rules:           []priority.Rule{},
				},
				Crawler: config.NewCrawlerConfig(),
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
