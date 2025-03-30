// Package config provides configuration functionality for the application.
package config

import "time"

// Interface defines the interface for configuration operations.
// It provides access to application configuration settings and
// supports different configuration sources.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *AppConfig
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *LogConfig
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetServerConfig returns the server configuration.
	GetServerConfig() *ServerConfig
	// GetSources returns the list of configured sources.
	GetSources() ([]Source, error)
	// GetCommand returns the current command being executed.
	GetCommand() string
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Environment string
	Name        string
	Version     string
	Debug       bool
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
	Debug bool
}

// ElasticsearchConfig holds Elasticsearch configuration.
type ElasticsearchConfig struct {
	Addresses []string
	IndexName string
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Address string
}

// Source holds configuration for a content source.
type Source struct {
	Name      string
	URL       string
	RateLimit time.Duration
	MaxDepth  int
	Time      []string
}

// Params holds the parameters for creating a config.
type Params struct {
	Environment string
	Debug       bool
	Command     string
}

// NewNoOp creates a no-op config that returns default values.
// This is useful for testing or when configuration is not needed.
func NewNoOp() Interface {
	return &NoOpConfig{}
}

// NoOpConfig implements Interface but returns default values.
type NoOpConfig struct{}

func (c *NoOpConfig) GetAppConfig() *AppConfig {
	return &AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       false,
	}
}

func (c *NoOpConfig) GetLogConfig() *LogConfig {
	return &LogConfig{
		Level: "info",
		Debug: false,
	}
}

func (c *NoOpConfig) GetElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "gocrawl",
	}
}

func (c *NoOpConfig) GetServerConfig() *ServerConfig {
	return &ServerConfig{
		Address: ":8080",
	}
}

func (c *NoOpConfig) GetSources() ([]Source, error) {
	return []Source{}, nil
}

func (c *NoOpConfig) GetCommand() string {
	return "test"
}
