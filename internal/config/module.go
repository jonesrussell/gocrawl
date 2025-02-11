package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Module provides the config as an Fx module
var Module = fx.Module("config",
	fx.Provide(
		func() (*Config, error) {
			// Don't require a config file
			viper.SetConfigName("config")
			viper.AddConfigPath(".")
			viper.AutomaticEnv()

			// Ignore error from reading config file since we'll use env vars
			_ = viper.ReadInConfig()

			// Set defaults
			viper.SetDefault("APP_ENV", "development")
			viper.SetDefault("LOG_LEVEL", "info")
			viper.SetDefault("ELASTIC_URL", "http://localhost:9200")
			viper.SetDefault("CRAWLER_RATE_LIMIT", "1s")
			viper.SetDefault("CRAWLER_MAX_DEPTH", 2)

			cfg := &Config{
				App: AppConfig{
					Environment: viper.GetString("APP_ENV"),
					LogLevel:    viper.GetString("LOG_LEVEL"),
					Debug:       viper.GetBool("APP_DEBUG"),
				},
				Crawler: CrawlerConfig{
					BaseURL:   viper.GetString("CRAWLER_BASE_URL"),
					MaxDepth:  viper.GetInt("CRAWLER_MAX_DEPTH"),
					RateLimit: viper.GetDuration("CRAWLER_RATE_LIMIT"),
					IndexName: viper.GetString("INDEX_NAME"),
				},
				Elasticsearch: ElasticsearchConfig{
					URL:      viper.GetString("ELASTIC_URL"),
					Password: viper.GetString("ELASTIC_PASSWORD"),
					APIKey:   viper.GetString("ELASTIC_API_KEY"),
				},
			}

			// Validate required fields
			if cfg.Elasticsearch.URL == "" {
				return nil, ErrMissingElasticURL
			}

			return cfg, nil
		},
	),
)
