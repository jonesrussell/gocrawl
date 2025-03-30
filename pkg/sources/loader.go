// Package sources provides source management functionality for the application.
package sources

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads source configurations from a file.
func LoadFromFile(path string) ([]Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
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
		return ErrInvalidSource
	}
	if cfg.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidSource)
	}
	if cfg.URL == "" {
		return fmt.Errorf("%w: URL is required", ErrInvalidSource)
	}
	if _, err := url.Parse(cfg.URL); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}
	if cfg.RateLimit < 0 {
		return fmt.Errorf("%w: rate limit must be positive", ErrInvalidRateLimit)
	}
	if cfg.MaxDepth < 0 {
		return fmt.Errorf("%w: max depth must be positive", ErrInvalidMaxDepth)
	}
	if len(cfg.Time) > 0 {
		for _, t := range cfg.Time {
			if _, err := time.Parse("15:04", t); err != nil {
				return fmt.Errorf("%w: invalid time format %s", ErrInvalidTime, t)
			}
		}
	}
	return nil
}
