package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

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

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "http://example.com", cfg.Crawler.BaseURL)
	assert.Equal(t, 3, cfg.Crawler.MaxDepth)
	assert.Equal(t, time.Second, cfg.Crawler.RateLimit)
	assert.Equal(t, "index", cfg.Crawler.IndexName)
	assert.Equal(t, "http://localhost:9200", cfg.Elasticsearch.URL)
	assert.Equal(t, "user", cfg.Elasticsearch.Username)
	assert.Equal(t, "pass", cfg.Elasticsearch.Password)
	assert.Equal(t, "apikey", cfg.Elasticsearch.APIKey)
	assert.Equal(t, "index", cfg.Elasticsearch.IndexName)
	assert.True(t, cfg.Elasticsearch.SkipTLS)
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

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, config.ErrMissingElasticURL, err)
}

func TestParseRateLimit(t *testing.T) {
	rateLimit, err := config.ParseRateLimit("1s")

	assert.NoError(t, err)
	assert.Equal(t, time.Second, rateLimit)

	// Test invalid duration
	rateLimit, err = config.ParseRateLimit("invalid")
	assert.NoError(t, err)
	assert.Equal(t, time.Second, rateLimit) // Should return default value
}

func TestValidateConfig(t *testing.T) {
	cfg := &config.Config{
		Elasticsearch: config.ElasticsearchConfig{
			URL: "http://localhost:9200",
		},
	}

	err := config.ValidateConfig(cfg)
	assert.NoError(t, err)

	// Test missing Elasticsearch URL
	cfg.Elasticsearch.URL = ""
	err = config.ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Equal(t, config.ErrMissingElasticURL, err)
}
