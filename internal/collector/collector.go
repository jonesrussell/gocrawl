// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
//
// The package is organized into several key components:
//
// 1. Collector: The main component that manages the crawling process
//   - Handles URL validation and processing
//   - Manages rate limiting and concurrency
//   - Configures crawler behavior
//
// 2. Content Processing: Handles different types of content
//   - Article detection and processing
//   - General content processing
//   - Link following and URL management
//
// 3. Context Management: Manages crawler state
//   - Stores HTML elements
//   - Tracks article status
//   - Maintains crawler context
//
// Usage example:
//
//	params := collector.Params{
//		BaseURL:   "https://example.com",
//		MaxDepth:  2,
//		RateLimit: time.Second,
//		Logger:    logger,
//	}
//	result, err := collector.New(params)
//	if err != nil {
//		log.Fatal(err)
//	}
//	result.Collector.Visit("https://example.com")
//
// The package follows these design principles:
// - Separation of concerns between crawling and content processing
// - Configurable behavior through parameters
// - Extensible processor interface for different content types
// - Robust error handling and logging
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
// 6. Sets up event handlers using the provided completion channel
//
// Parameters:
//   - p: Params containing all required configuration and dependencies
//
// Returns:
//   - Result: Contains the configured collector
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

	// Use the provided done channel for event handlers
	handlers := NewHandlers(cfg, p.Done, c)
	handlers.ConfigureHandlers()

	// Log successful collector creation with configuration details
	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	// Return configured collector
	return Result{Collector: c}, nil
}
