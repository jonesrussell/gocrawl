package collector

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Setup handles the setup and configuration of the collector
type Setup struct {
	config *Config
}

// NewSetup creates a new collector setup instance
func NewSetup(config *Config) *Setup {
	return &Setup{
		config: config,
	}
}

// CreateBaseCollector creates a new collector with base configuration
func (s *Setup) CreateBaseCollector(domain string) *colly.Collector {
	return colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(s.config.MaxDepth),
		colly.AllowedDomains(domain),
	)
}

// ConfigureCollector sets up the collector with all necessary settings
func (s *Setup) ConfigureCollector(c *colly.Collector) error {
	// Configure rate limiting
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: s.config.RandomDelay,
		Parallelism: s.config.Parallelism,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Configure debugger and logging
	if s.config.Debugger != nil {
		c.SetDebugger(s.config.Debugger)
	}
	ConfigureLogging(c, s.config.Logger)

	// Configure content processing
	contentParams := ContentParams{
		Logger:           s.config.Logger,
		Source:           s.config.Source,
		ArticleProcessor: s.config.ArticleProcessor,
		ContentProcessor: s.config.ContentProcessor,
	}
	configureContentProcessing(c, contentParams)

	return nil
}

// ValidateURL validates the base URL
func (s *Setup) ValidateURL() error {
	parsedURL, err := url.Parse(s.config.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", s.config.BaseURL)
	}
	return nil
}
