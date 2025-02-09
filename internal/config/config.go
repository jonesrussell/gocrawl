package config

import (
	"os"

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
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err // Handle error if .env file is not found
	}

	appDebug := os.Getenv("APP_DEBUG") == "true"

	return &Config{
		AppName:         os.Getenv("APP_NAME"),
		AppEnv:          os.Getenv("APP_ENV"),
		AppDebug:        appDebug,
		ElasticURL:      os.Getenv("ELASTIC_URL"),
		ElasticPassword: os.Getenv("ELASTIC_PASSWORD"),
		ElasticAPIKey:   os.Getenv("ELASTIC_API_KEY"),
	}, nil
}
