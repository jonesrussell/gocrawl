// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// createElasticsearchConfig creates a new ElasticsearchConfig from Viper values
func createElasticsearchConfig(v *viper.Viper) *ElasticsearchConfig {
	// Debug: Print values being set
	fmt.Printf("\nCreating Elasticsearch config:\n")
	fmt.Printf("Addresses: %v\n", v.GetStringSlice("elasticsearch.addresses"))
	fmt.Printf("Username: %s\n", v.GetString("elasticsearch.username"))
	fmt.Printf("Password: %s\n", v.GetString("elasticsearch.password"))
	fmt.Printf("API Key: %s\n", v.GetString("elasticsearch.api_key"))
	fmt.Printf("Index Name: %s\n", v.GetString("elasticsearch.index_name"))
	fmt.Printf("Cloud ID: %s\n", v.GetString("elasticsearch.cloud.id"))
	fmt.Printf("Cloud API Key: %s\n", v.GetString("elasticsearch.cloud.api_key"))
	fmt.Printf("TLS Enabled: %v\n", v.GetBool("elasticsearch.tls.enabled"))
	fmt.Printf("TLS Certificate: %s\n", v.GetString("elasticsearch.tls.certificate"))
	fmt.Printf("TLS Key: %s\n", v.GetString("elasticsearch.tls.key"))
	fmt.Printf("Retry Enabled: %v\n", v.GetBool("elasticsearch.retry.enabled"))
	fmt.Printf("Retry Initial Wait: %s\n", v.GetString("elasticsearch.retry.initial_wait"))
	fmt.Printf("Retry Max Wait: %s\n", v.GetString("elasticsearch.retry.max_wait"))
	fmt.Printf("Retry Max Retries: %d\n", v.GetInt("elasticsearch.retry.max_retries"))

	config := &ElasticsearchConfig{
		Addresses: v.GetStringSlice("elasticsearch.addresses"),
		Username:  v.GetString("elasticsearch.username"),
		Password:  v.GetString("elasticsearch.password"),
		APIKey:    v.GetString("elasticsearch.api_key"),
		IndexName: v.GetString("elasticsearch.index_name"),
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{
			ID:     v.GetString("elasticsearch.cloud.id"),
			APIKey: v.GetString("elasticsearch.cloud.api_key"),
		},
		TLS: TLSConfig{
			Enabled:  v.GetBool("elasticsearch.tls.enabled"),
			CertFile: v.GetString("elasticsearch.tls.certificate"),
			KeyFile:  v.GetString("elasticsearch.tls.key"),
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     v.GetBool("elasticsearch.retry.enabled"),
			InitialWait: v.GetDuration("elasticsearch.retry.initial_wait"),
			MaxWait:     v.GetDuration("elasticsearch.retry.max_wait"),
			MaxRetries:  v.GetInt("elasticsearch.retry.max_retries"),
		},
	}

	// Debug: Print final config
	fmt.Printf("\nFinal Elasticsearch config:\n")
	fmt.Printf("%+v\n", config)
	return config
}
