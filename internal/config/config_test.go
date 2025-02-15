package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestNewConfig(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.ReadConfig(strings.NewReader(`
APP_ENV: "development"
LOG_LEVEL: "debug"
APP_DEBUG: true
CRAWLER_BASE_URL: "http://example.com"
CRAWLER_MAX_DEPTH: 3
CRAWLER_RATE_LIMIT: "1s"
ELASTIC_URL: "http://localhost:9200"
ELASTIC_USERNAME: "user"
ELASTIC_PASSWORD: "pass"
ELASTIC_API_KEY: "apikey"
ELASTIC_INDEX_NAME: "index"
ELASTIC_SKIP_TLS: true
    `))

	cfg, err := config.NewConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "development", cfg.App.Environment)
	require.Equal(t, "debug", cfg.App.LogLevel)
	require.True(t, cfg.App.Debug)
	require.Equal(t, "http://example.com", cfg.Crawler.BaseURL)
	require.Equal(t, 3, cfg.Crawler.MaxDepth)
	require.Equal(t, time.Second, cfg.Crawler.RateLimit)
	require.Equal(t, "index", cfg.Crawler.IndexName)
	require.Equal(t, "http://localhost:9200", cfg.Elasticsearch.URL)
	require.Equal(t, "user", cfg.Elasticsearch.Username)
	require.Equal(t, "pass", cfg.Elasticsearch.Password)
	require.Equal(t, "apikey", cfg.Elasticsearch.APIKey)
	require.Equal(t, "index", cfg.Elasticsearch.IndexName)
	require.True(t, cfg.Elasticsearch.SkipTLS)
}

func TestNewConfig_MissingElasticURL(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.ReadConfig(strings.NewReader(`
APP_ENV: "development"
LOG_LEVEL: "debug"
APP_DEBUG: true
CRAWLER_BASE_URL: "http://example.com"
CRAWLER_MAX_DEPTH: 3
CRAWLER_RATE_LIMIT: "1s"
ELASTIC_USERNAME: "user"
ELASTIC_PASSWORD: "pass"
ELASTIC_API_KEY: "apikey"
ELASTIC_INDEX_NAME: "index"
ELASTIC_SKIP_TLS: true
    `))

	cfg, err := config.NewConfig()

	require.Error(t, err)
	require.Nil(t, cfg)
	require.Equal(t, config.ErrMissingElasticURL, err)
}

func TestParseRateLimit(t *testing.T) {
	rateLimit, err := config.ParseRateLimit("1s")

	require.NoError(t, err)
	require.Equal(t, time.Second, rateLimit)

	// Test invalid duration
	rateLimit, err = config.ParseRateLimit("invalid")
	require.NoError(t, err)
	require.Equal(t, time.Second, rateLimit) // Should return default value
}

func TestValidateConfig(t *testing.T) {
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
	}

	err := config.ValidateConfig(cfg)
	require.NoError(t, err)

	// Test missing Elasticsearch URL
	cfg.Elasticsearch.URL = ""
	err = config.ValidateConfig(cfg)
	require.Error(t, err)
	require.Equal(t, config.ErrMissingElasticURL, err)
}
