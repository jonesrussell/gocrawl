// Package sources provides source management functionality for the application.
package sources

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the configuration file structure.
type FileConfig struct {
	Sources []struct {
		Name      string        `yaml:"name"`
		URL       string        `yaml:"url"`
		RateLimit time.Duration `yaml:"rate_limit"`
		MaxDepth  int           `yaml:"max_depth"`
		Time      []string      `yaml:"time"`
	} `yaml:"sources"`
}

// LoadFromFile loads source configurations from a YAML file.
func LoadFromFile(path string) ([]Config, error) {
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, fmt.Errorf("failed to open file: %w", readErr)
	}

	var configs []Config
	if unmarshalErr := yaml.Unmarshal(data, &configs); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", unmarshalErr)
	}

	// Validate each config
	for i, cfg := range configs {
		if validateErr := ValidateConfig(&cfg); validateErr != nil {
			return nil, fmt.Errorf("invalid config at index %d: %w", i, validateErr)
		}
	}

	return configs, nil
}

// ValidateConfig validates a source configuration.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	if cfg.Name == "" {
		return errors.New("name is required")
	}

	if cfg.URL == "" {
		return errors.New("URL is required")
	}

	// Parse URL and validate scheme
	u, parseErr := url.Parse(cfg.URL)
	if parseErr != nil || u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("invalid URL: must be a valid HTTP(S) URL")
	}

	if cfg.RateLimit <= 0 {
		return errors.New("rate limit must be positive")
	}

	if cfg.MaxDepth <= 0 {
		return errors.New("max depth must be positive")
	}

	// Validate time format if provided
	if len(cfg.Time) > 0 {
		for _, t := range cfg.Time {
			if _, timeErr := time.Parse("15:04", t); timeErr != nil {
				return fmt.Errorf("invalid time format: %w", timeErr)
			}
		}
	}

	return nil
}
