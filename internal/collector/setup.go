// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Setup handles the setup and configuration of the collector.
// It manages the creation and configuration of the Colly collector instance,
// including rate limiting, debugging, logging, and content processing.
type Setup struct {
	// config contains the collector configuration
	config *Config
}

// NewSetup creates a new collector setup instance.
// It initializes the setup with the provided configuration.
//
// Parameters:
//   - config: The collector configuration
//
// Returns:
//   - *Setup: The initialized setup instance
func NewSetup(config *Config) *Setup {
	return &Setup{
		config: config,
	}
}

// CreateBaseCollector creates a new collector with base configuration.
// It sets up the collector with basic settings including:
// - Asynchronous operation
// - Maximum depth limit
// - Domain restrictions
//
// Parameters:
//   - domain: The domain to restrict crawling to
//
// Returns:
//   - *colly.Collector: The configured collector instance
func (s *Setup) CreateBaseCollector(domain string) *colly.Collector {
	return colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(s.config.MaxDepth),
		colly.AllowedDomains(domain),
	)
}

// ConfigureCollector sets up the collector with all necessary settings.
// It configures:
// - Rate limiting
// - Debugging
// - Logging
// - Content processing
//
// Parameters:
//   - c: The collector instance to configure
//
// Returns:
//   - error: Any error that occurred during configuration
func (s *Setup) ConfigureCollector(c *colly.Collector) error {
	// Configure rate limiting with domain-specific rules
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: s.config.RandomDelay,
		Parallelism: s.config.Parallelism,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Configure debugger and logging if enabled
	if s.config.Debugger != nil {
		c.SetDebugger(s.config.Debugger)
	}
	ConfigureLogging(c, s.config.Logger)

	// Configure content processing with provided processors
	contentParams := ContentParams{
		Logger:           s.config.Logger,
		Source:           s.config.Source,
		ArticleProcessor: s.config.ArticleProcessor,
		ContentProcessor: s.config.ContentProcessor,
	}
	configureContentProcessing(c, contentParams)

	return nil
}

// ValidateURL validates the base URL.
// It ensures the URL is:
// - Well-formed
// - Uses HTTP or HTTPS protocol
//
// Returns:
//   - error: Any validation error that occurred
func (s *Setup) ValidateURL() error {
	parsedURL, err := url.Parse(s.config.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", s.config.BaseURL)
	}
	return nil
}
