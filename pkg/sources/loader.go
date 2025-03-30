// Package sources provides source management functionality for the application.
package sources

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// LoadFromFile loads source configurations from a file.
func LoadFromFile(path string) ([]Config, error) {
	// TODO: Implement file loading
	return nil, errors.New("file loading not implemented")
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
