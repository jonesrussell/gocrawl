// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/types"
)

const (
	defaultMaxRetries = 3
)

// ElasticsearchConfig represents Elasticsearch configuration settings.
type ElasticsearchConfig struct {
	// Addresses is a list of Elasticsearch node addresses
	Addresses []string `yaml:"addresses"`
	// APIKey is the API key for authentication
	APIKey string `yaml:"api_key"`
	// Username is the username for authentication
	Username string `yaml:"username"`
	// Password is the password for authentication
	Password string `yaml:"password"`
	// IndexName is the name of the index
	IndexName string `yaml:"index_name"`
	// Cloud contains cloud-specific configuration
	Cloud struct {
		ID     string `yaml:"id"`
		APIKey string `yaml:"api_key"`
	} `yaml:"cloud"`
	// TLS contains TLS configuration
	TLS *types.TLSConfig `yaml:"tls"`
	// Retry contains retry configuration
	Retry struct {
		Enabled     bool          `yaml:"enabled"`
		InitialWait time.Duration `yaml:"initial_wait"`
		MaxWait     time.Duration `yaml:"max_wait"`
		MaxRetries  int           `yaml:"max_retries"`
	} `yaml:"retry"`
}

// NewElasticsearchConfig creates a new ElasticsearchConfig with default values.
func NewElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{},
		TLS: &types.TLSConfig{
			Enabled: false,
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     true,
			InitialWait: time.Second,
			MaxWait:     time.Minute,
			MaxRetries:  defaultMaxRetries,
		},
	}
}

// Validate validates the Elasticsearch configuration.
func (c *ElasticsearchConfig) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("at least one address is required")
	}
	if c.TLS != nil {
		if err := c.TLS.Validate(); err != nil {
			return fmt.Errorf("tls: %w", err)
		}
	}
	return nil
}
