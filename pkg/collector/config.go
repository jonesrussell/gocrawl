// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Error messages used for parameter validation
const (
	// errEmptyBaseURL is returned when the base URL is not provided
	errEmptyBaseURL = "base URL is required"
	// errMissingArticleProc is returned when the article processor is not provided
	errMissingArticleProc = "article processor is required"
	// errMissingLogger is returned when the logger is not provided
	errMissingLogger = "logger is required"
	// errMissingDone is returned when the done channel is not provided
	errMissingDone = "done channel is required"
	// errMissingSource is returned when the source configuration is not provided
	errMissingSource = "source configuration is required"
)

// Params holds the parameters for creating a Collector.
// It uses fx.In for dependency injection and contains all required
// configuration and dependencies for the collector.
type Params struct {
	fx.In

	// ArticleProcessor handles the processing of article content
	ArticleProcessor Processor
	// ContentProcessor handles the processing of general content
	ContentProcessor Processor
	// BaseURL is the starting URL for crawling
	BaseURL string
	// Context provides cancellation and timeout support
	Context context.Context
	// Debugger handles debugging operations
	Debugger debug.Debugger
	// Logger provides logging capabilities
	Logger logger.Interface
	// MaxDepth is the maximum crawling depth
	MaxDepth int
	// Parallelism is the number of concurrent requests
	Parallelism int
	// RandomDelay adds random delay between requests
	RandomDelay time.Duration
	// RateLimit is the time between requests
	RateLimit time.Duration
	// Source contains source-specific configuration
	Source *config.Source
	// Done is a channel that signals when crawling is complete
	Done chan struct{} `name:"crawlDone"`
	// AllowedDomains is a list of domains that the collector is allowed to visit
	AllowedDomains []string
}

// Result holds the collector instance.
// It uses fx.Out for dependency injection.
type Result struct {
	fx.Out

	// Collector is the configured Colly collector instance
	Collector *colly.Collector
}

// ValidateParams validates the collector parameters.
// It ensures all required fields are provided and valid.
//
// Parameters:
//   - p: Params to validate
//
// Returns:
//   - error: Any validation error that occurred
func ValidateParams(p Params) error {
	// Ensure BaseURL is not empty
	if p.BaseURL == "" {
		return errors.New(errEmptyBaseURL)
	}

	// Ensure ArticleProcessor is provided
	if p.ArticleProcessor == nil {
		return errors.New(errMissingArticleProc)
	}

	// Ensure Logger is provided
	if p.Logger == nil {
		return errors.New(errMissingLogger)
	}

	// Ensure Done channel is provided
	if p.Done == nil {
		return errors.New(errMissingDone)
	}

	// Ensure Source is provided
	if p.Source == nil {
		return errors.New(errMissingSource)
	}

	return nil
}

// Config holds the configuration for the collector.
type Config struct {
	// Logger is the logger for the collector.
	Logger logger.Interface

	// MaxDepth is the maximum depth to crawl.
	MaxDepth int

	// RateLimit is the rate limit for requests.
	RateLimit time.Duration

	// Processors are the processors to use.
	Processors []Processor

	// Debugger is the debugger to use.
	Debugger debug.Debugger

	// BaseURL is the starting URL for crawling.
	BaseURL string

	// Parallelism is the number of concurrent requests.
	Parallelism int

	// Source contains source-specific configuration.
	Source config.Source

	// ArticleProcessor handles the processing of article content.
	ArticleProcessor Processor

	// ContentProcessor handles the processing of general content.
	ContentProcessor Processor
}

// NewConfig creates a new Config instance with the given parameters.
//
// Parameters:
//   - p: Params to create the config from
//
// Returns:
//   - *Config: The new config instance
//   - error: Any error that occurred during creation
func NewConfig(p Params) (*Config, error) {
	// Validate parameters
	if err := ValidateParams(p); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Create config
	config := &Config{
		Logger:           p.Logger,
		MaxDepth:         p.MaxDepth,
		RateLimit:        p.RateLimit,
		Processors:       []Processor{p.ArticleProcessor, p.ContentProcessor},
		Debugger:         p.Debugger,
		BaseURL:          p.BaseURL,
		Parallelism:      p.Parallelism,
		Source:           *p.Source,
		ArticleProcessor: p.ArticleProcessor,
		ContentProcessor: p.ContentProcessor,
	}

	// Validate config
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// ValidateConfig validates the collector configuration.
//
// Returns:
//   - error: Any validation error that occurred
func (c *Config) ValidateConfig() error {
	// Ensure Logger is provided
	if c.Logger == nil {
		return errors.New(errMissingLogger)
	}

	// Ensure at least one processor is provided
	if len(c.Processors) == 0 {
		return errors.New("at least one processor is required")
	}

	return nil
}
