// Package sources provides source management functionality for the application.
package sources

import (
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	var configs []Config
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate each config
	for i, cfg := range configs {
		if err := ValidateConfig(&cfg); err != nil {
			return nil, fmt.Errorf("invalid config at index %d: %w", i, err)
		}
	}

	return configs, nil
}

// ValidateConfig validates a source configuration.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}

	if cfg.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if _, err := url.Parse(cfg.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if cfg.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}

	if cfg.MaxDepth <= 0 {
		return fmt.Errorf("max depth must be positive")
	}

	// Validate time format if provided
	if len(cfg.Time) > 0 {
		for _, t := range cfg.Time {
			if _, err := time.Parse("15:04", t); err != nil {
				return fmt.Errorf("invalid time format: %w", err)
			}
		}
	}

	return nil
}
