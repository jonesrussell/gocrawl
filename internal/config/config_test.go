package config_test

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestNewConfig(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")     // Name of the file without extension
	viper.AddConfigPath("./testdata") // Path to the testdata directory

	err := viper.ReadInConfig()
	require.NoError(t, err)

	cfg, err := config.NewConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test", cfg.App.Environment)
	require.Equal(t, "debug", cfg.Log.Level)
	require.True(t, cfg.Log.Debug)
	require.Equal(t, "http://test.example.com", cfg.Crawler.BaseURL)
	require.Equal(t, 5, cfg.Crawler.MaxDepth)
	require.Equal(t, 2*time.Second, cfg.Crawler.RateLimit)
	require.Equal(t, "http://localhost:9200", cfg.Elasticsearch.URL)
	require.Equal(t, "test_user", cfg.Elasticsearch.Username)
	require.Equal(t, "test_pass", cfg.Elasticsearch.Password)
	require.Equal(t, "test_apikey", cfg.Elasticsearch.APIKey)
	require.Equal(t, "test_index", cfg.Elasticsearch.IndexName)
	require.False(t, cfg.Elasticsearch.SkipTLS)
}

func TestParseRateLimit(t *testing.T) {
	rateLimit, err := config.ParseRateLimit("1s")
	require.NoError(t, err)
	require.Equal(t, time.Second, rateLimit)

	// Test invalid duration
	rateLimit, err = config.ParseRateLimit("invalid")
	require.Error(t, err)                    // Expect an error
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
