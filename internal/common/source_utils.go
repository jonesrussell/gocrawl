// Package common provides shared utilities and types used across the application.
package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

// ConvertSourceConfig converts a sources.Config to a config.Source.
// It handles the conversion of fields between the two types.
func ConvertSourceConfig(source *sources.Config) *config.Source {
	if source == nil {
		return nil
	}

	return &config.Source{
		Name:      source.Name,
		URL:       source.URL,
		RateLimit: source.RateLimit,
		MaxDepth:  source.MaxDepth,
		Time:      source.Time,
	}
}

// ExtractDomain extracts the domain from a URL string.
// It handles both full URLs and path-only URLs.
func ExtractDomain(sourceURL string) (string, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Host == "" {
		// If no host in URL, treat the first path segment as the domain
		parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(parts) > 0 {
			return parts[0], nil
		}
		return "", fmt.Errorf("could not extract domain from path: %s", sourceURL)
	}

	return parsedURL.Host, nil
}
