// Package config provides configuration management for the GoCrawl application.
package config

import (
	"time"

	"github.com/spf13/viper"
)

// createElasticsearchConfig creates the Elasticsearch configuration
func createElasticsearchConfig() ElasticsearchConfig {
	// Get retry settings with defaults
	retryEnabled := viper.GetBool("elasticsearch.retry.enabled")
	retryInitialWait := viper.GetDuration("elasticsearch.retry.initial_wait")
	if retryInitialWait == 0 {
		retryInitialWait = defaultRetryInitialWait
	}
	retryMaxWait := viper.GetDuration("elasticsearch.retry.max_wait")
	if retryMaxWait == 0 {
		retryMaxWait = defaultRetryMaxWait
	}
	retryMaxRetries := viper.GetInt("elasticsearch.retry.max_retries")
	if retryMaxRetries == 0 {
		retryMaxRetries = defaultMaxRetries
	}

	return ElasticsearchConfig{
		Addresses: viper.GetStringSlice("elasticsearch.addresses"),
		Username:  viper.GetString("elasticsearch.username"),
		Password:  viper.GetString("elasticsearch.password"),
		APIKey:    viper.GetString("elasticsearch.api_key"),
		IndexName: viper.GetString("elasticsearch.index_name"),
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{
			ID:     viper.GetString("elasticsearch.cloud.id"),
			APIKey: viper.GetString("elasticsearch.cloud.api_key"),
		},
		TLS: TLSConfig{
			Enabled:  viper.GetBool("elasticsearch.tls.enabled"),
			CertFile: viper.GetString("elasticsearch.tls.certificate"),
			KeyFile:  viper.GetString("elasticsearch.tls.key"),
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     retryEnabled,
			InitialWait: retryInitialWait,
			MaxWait:     retryMaxWait,
			MaxRetries:  retryMaxRetries,
		},
	}
}
