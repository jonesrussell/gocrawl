// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
package sources

import (
	"fmt"
	"net/url"
	"strings"

	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	sourcestypes "github.com/jonesrussell/gocrawl/internal/sources/types"
)

// ConvertSourceConfig converts a sourcestypes.SourceConfig to a configtypes.Source.
// It handles the conversion of fields between the two types.
func ConvertSourceConfig(source *sourcestypes.SourceConfig) *configtypes.Source {
	if source == nil {
		return nil
	}

	return sourcestypes.ConvertToConfigSource(source)
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
