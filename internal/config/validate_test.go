package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Elasticsearch: config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
				},
				Crawler: config.CrawlerConfig{
					Parallelism: 2,
					MaxDepth:    2,
					RateLimit:   time.Second,
					RandomDelay: time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid app environment",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid crawler max depth",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Crawler: config.CrawlerConfig{
					MaxDepth: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid crawler parallelism",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Crawler: config.CrawlerConfig{
					Parallelism: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "server security enabled without API key",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Crawler: config.CrawlerConfig{
					MaxDepth:    2,
					Parallelism: 2,
				},
				Server: config.ServerConfig{
					Security: struct {
						Enabled   bool   `yaml:"enabled"`
						APIKey    string `yaml:"api_key"`
						RateLimit int    `yaml:"rate_limit"`
						CORS      struct {
							Enabled        bool     `yaml:"enabled"`
							AllowedOrigins []string `yaml:"allowed_origins"`
							AllowedMethods []string `yaml:"allowed_methods"`
							AllowedHeaders []string `yaml:"allowed_headers"`
							MaxAge         int      `yaml:"max_age"`
						} `yaml:"cors"`
						TLS struct {
							Enabled     bool   `yaml:"enabled"`
							Certificate string `yaml:"certificate"`
							Key         string `yaml:"key"`
						} `yaml:"tls"`
					}{
						Enabled: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "server security enabled with invalid API key",
			config: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Crawler: config.CrawlerConfig{
					MaxDepth:    2,
					Parallelism: 2,
				},
				Server: config.ServerConfig{
					Security: struct {
						Enabled   bool   `yaml:"enabled"`
						APIKey    string `yaml:"api_key"`
						RateLimit int    `yaml:"rate_limit"`
						CORS      struct {
							Enabled        bool     `yaml:"enabled"`
							AllowedOrigins []string `yaml:"allowed_origins"`
							AllowedMethods []string `yaml:"allowed_methods"`
							AllowedHeaders []string `yaml:"allowed_headers"`
							MaxAge         int      `yaml:"max_age"`
						} `yaml:"cors"`
						TLS struct {
							Enabled     bool   `yaml:"enabled"`
							Certificate string `yaml:"certificate"`
							Key         string `yaml:"key"`
						} `yaml:"tls"`
					}{
						Enabled: true,
						APIKey:  "invalid-key",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
