// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"fmt"
	"net/url"
)

// New creates a new collector instance with the specified parameters.
// It performs the following steps:
// 1. Validates the input parameters
// 2. Creates and validates the collector configuration
// 3. Sets up the collector with the configuration
// 4. Validates the base URL
// 5. Creates and configures the base collector
// 6. Sets up event handlers and completion channel
//
// Parameters:
//   - p: Params containing all required configuration and dependencies
//
// Returns:
//   - Result: Contains the configured collector and completion channel
//   - error: Any error that occurred during setup
func New(p Params) (Result, error) {
	// Validate input parameters
	if err := ValidateParams(p); err != nil {
		return Result{}, err
	}

	// Create collector configuration from parameters
	cfg, err := NewConfig(p)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create collector config: %w", err)
	}

	// Validate the configuration
	if validateErr := cfg.ValidateConfig(); validateErr != nil {
		return Result{}, fmt.Errorf("invalid collector config: %w", validateErr)
	}

	// Create and validate collector setup
	setup := NewSetup(cfg)
	if urlErr := setup.ValidateURL(); urlErr != nil {
		return Result{}, urlErr
	}

	// Parse URL to extract domain for collector configuration
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return Result{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create base collector with domain restrictions
	c := setup.CreateBaseCollector(parsedURL.Hostname())

	// Configure collector with all settings
	if configErr := setup.ConfigureCollector(c); configErr != nil {
		return Result{}, fmt.Errorf("failed to configure collector: %w", configErr)
	}

	// Create completion channel and event handlers
	done := make(chan struct{})
	handlers := NewHandlers(cfg, done)
	handlers.ConfigureHandlers(c)

	// Log successful collector creation with configuration details
	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	// Return configured collector and completion channel
	return Result{Collector: c, Done: done}, nil
}
