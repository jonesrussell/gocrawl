// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"net/url"
	"strings"
)

// validateHTTPMethod checks if the given HTTP method is valid.
func validateHTTPMethod(method string) bool {
	return ValidHTTPMethods[method]
}

// validateHTTPHeaders checks if all given HTTP headers are valid.
func validateHTTPHeaders(headers map[string]string) bool {
	for header := range headers {
		if !ValidHTTPHeaders[header] {
			return false
		}
	}
	return true
}

// validateLogLevel checks if the given log level is valid.
func validateLogLevel(level string) bool {
	return ValidLogLevels[level]
}

// validateEnvironment checks if the given environment is valid.
func validateEnvironment(env string) bool {
	return ValidEnvironments[env]
}

// validateURL checks if the given URL is valid.
func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return fmt.Errorf("URL must use HTTP(S) scheme")
	}

	return nil
}

// validateURLs checks if all given URLs are valid.
func validateURLs(urls []string) error {
	for _, u := range urls {
		if err := validateURL(u); err != nil {
			return fmt.Errorf("invalid URL %q: %w", u, err)
		}
	}
	return nil
}

// validateSelector checks if the given selector is valid.
func validateSelector(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}
	return nil
}

// validateSelectors checks if all required selectors are present and valid.
func validateSelectors(selectors map[string]string, required []string) error {
	for _, req := range required {
		sel, ok := selectors[req]
		if !ok {
			return fmt.Errorf("missing required selector %q", req)
		}
		if err := validateSelector(sel); err != nil {
			return fmt.Errorf("invalid selector %q: %w", req, err)
		}
	}
	return nil
}
