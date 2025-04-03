// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// createElasticsearchConfig creates a new ElasticsearchConfig from Viper values
func createElasticsearchConfig() *ElasticsearchConfig {
	// Debug: Print values being set
	fmt.Printf("\nCreating Elasticsearch config:\n")
	fmt.Printf("Addresses: %v\n", viper.GetStringSlice("elasticsearch.addresses"))
	fmt.Printf("Username: %s\n", viper.GetString("elasticsearch.username"))
	fmt.Printf("Password: %s\n", viper.GetString("elasticsearch.password"))
	fmt.Printf("API Key: %s\n", viper.GetString("elasticsearch.api_key"))
	fmt.Printf("Index Name: %s\n", viper.GetString("elasticsearch.index_name"))
	fmt.Printf("Cloud ID: %s\n", viper.GetString("elasticsearch.cloud.id"))
	fmt.Printf("Cloud API Key: %s\n", viper.GetString("elasticsearch.cloud.api_key"))
	fmt.Printf("TLS Enabled: %v\n", viper.GetBool("elasticsearch.tls.enabled"))
	fmt.Printf("TLS Certificate: %s\n", viper.GetString("elasticsearch.tls.certificate"))
	fmt.Printf("TLS Key: %s\n", viper.GetString("elasticsearch.tls.key"))
	fmt.Printf("Retry Enabled: %v\n", viper.GetBool("elasticsearch.retry.enabled"))
	fmt.Printf("Retry Initial Wait: %s\n", viper.GetString("elasticsearch.retry.initial_wait"))
	fmt.Printf("Retry Max Wait: %s\n", viper.GetString("elasticsearch.retry.max_wait"))
	fmt.Printf("Retry Max Retries: %d\n", viper.GetInt("elasticsearch.retry.max_retries"))

	config := &ElasticsearchConfig{
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
			Enabled:     viper.GetBool("elasticsearch.retry.enabled"),
			InitialWait: viper.GetDuration("elasticsearch.retry.initial_wait"),
			MaxWait:     viper.GetDuration("elasticsearch.retry.max_wait"),
			MaxRetries:  viper.GetInt("elasticsearch.retry.max_retries"),
		},
	}

	// Debug: Print final config
	fmt.Printf("\nFinal Elasticsearch config:\n")
	fmt.Printf("%+v\n", config)
	return config
}
