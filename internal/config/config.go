package config

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"
)

// CrawlerConfig holds crawler-specific configuration
type CrawlerConfig struct {
	BaseURL   string
	MaxDepth  int
	RateLimit time.Duration
}

// Config holds all configuration settings
type Config struct {
	// App settings
	AppEnv   string
	LogLevel string

	// Crawler settings
	IndexName string
	BaseURL   string
	MaxDepth  int

	// Elasticsearch settings
	ElasticURL      string
	ElasticPassword string
	ElasticAPIKey   string
	Transport       http.RoundTripper // For testing
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:        getEnvDefault("APP_ENV", "development"),
		LogLevel:      getEnvDefault("LOG_LEVEL", "info"),
		ElasticURL:    os.Getenv("ELASTIC_URL"),
		ElasticAPIKey: os.Getenv("ELASTIC_API_KEY"),
		IndexName:     getEnvDefault("INDEX_NAME", "articles"),
		BaseURL:       os.Getenv("CRAWLER_BASE_URL"),
		MaxDepth:      getIntEnvDefault("CRAWLER_MAX_DEPTH", 2),
	}

	// Validate required fields
	if cfg.ElasticURL == "" {
		return nil, errors.New("ELASTIC_URL is required")
	}

	return cfg, nil
}

// getEnvDefault returns the value of an environment variable or a default value if it's not set
func getEnvDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getIntEnvDefault returns the value of an integer environment variable or a default value if it's not set
func getIntEnvDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, _ := strconv.Atoi(value)
	return intValue
}
