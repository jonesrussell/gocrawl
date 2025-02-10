package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	AppName         string
	AppEnv          string
	AppDebug        bool
	ElasticURL      string
	ElasticPassword string
	ElasticAPIKey   string
	IndexName       string
	LogLevel        string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Fallback logging mechanism
		if os.Getenv("APP_ENV") == "development" {
			// Only print in development mode
			os.Stderr.WriteString("Warning: .env file not found, using environment variables\n")
		}
	}

	// Parse APP_DEBUG as a boolean
	appDebug, err := parseBoolEnv("APP_DEBUG", false)
	if err != nil {
		return nil, fmt.Errorf("error parsing APP_DEBUG: %w", err)
	}

	config := &Config{
		AppName:         os.Getenv("APP_NAME"),
		AppEnv:          os.Getenv("APP_ENV"),
		AppDebug:        appDebug,
		ElasticURL:      os.Getenv("ELASTIC_URL"),
		ElasticPassword: os.Getenv("ELASTIC_PASSWORD"),
		ElasticAPIKey:   os.Getenv("ELASTIC_API_KEY"),
		IndexName:       os.Getenv("INDEX_NAME"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
	}

	// Validate required configuration values
	if config.ElasticURL == "" {
		return nil, errors.New("ELASTIC_URL is required")
	}

	return config, nil
}

// parseBoolEnv parses a boolean environment variable with a default value
func parseBoolEnv(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(value)
}
