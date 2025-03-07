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
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Error messages used for parameter validation
const (
	// errEmptyBaseURL is returned when the base URL is not provided
	errEmptyBaseURL = "base URL cannot be empty"
	// errMissingArticleProc is returned when the article processor is not provided
	errMissingArticleProc = "article processor is required"
	// errMissingLogger is returned when the logger is not provided
	errMissingLogger = "logger is required"
)

// Params holds the parameters for creating a Collector.
// It uses fx.In for dependency injection and contains all required
// configuration and dependencies for the collector.
type Params struct {
	fx.In

	// ArticleProcessor handles the processing of article content
	ArticleProcessor models.ContentProcessor
	// ContentProcessor handles the processing of general content
	ContentProcessor models.ContentProcessor
	// BaseURL is the starting URL for crawling
	BaseURL string
	// Context provides cancellation and timeout support
	Context context.Context
	// Debugger handles debugging operations
	Debugger *logger.CollyDebugger
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
	Source *sources.Config
}

// Result holds the collector instance and completion channel.
// It uses fx.Out for dependency injection.
type Result struct {
	fx.Out

	// Collector is the configured Colly collector instance
	Collector *colly.Collector
	// Done is a channel that signals when crawling is complete
	Done chan struct{}
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

	return nil
}

// Config holds all configuration for the collector.
// It combines source-specific settings with general collector settings.
type Config struct {
	// BaseURL is the starting URL for crawling
	BaseURL string
	// MaxDepth is the maximum crawling depth
	MaxDepth int
	// RateLimit is the time between requests as a string
	RateLimit string
	// Parallelism is the number of concurrent requests
	Parallelism int
	// RandomDelay adds random delay between requests
	RandomDelay time.Duration
	// Debugger handles debugging operations
	Debugger *logger.CollyDebugger
	// Logger provides logging capabilities
	Logger logger.Interface
	// Source contains source-specific configuration
	Source config.Source
	// ArticleProcessor handles the processing of article content
	ArticleProcessor models.ContentProcessor
	// ContentProcessor handles the processing of general content
	ContentProcessor models.ContentProcessor
}

// NewConfig creates a new collector configuration from the provided parameters.
// It parses the rate limit duration and creates a complete configuration.
//
// Parameters:
//   - p: Params containing all required configuration
//
// Returns:
//   - *Config: The created configuration
//   - error: Any error that occurred during creation
func NewConfig(p Params) (*Config, error) {
	// Parse rate limit duration from string
	rateLimit, err := time.ParseDuration(p.Source.RateLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit: %w", err)
	}

	// Create and return new configuration
	return &Config{
		BaseURL:          p.BaseURL,
		MaxDepth:         p.MaxDepth,
		RateLimit:        p.Source.RateLimit,
		Parallelism:      p.Parallelism,
		RandomDelay:      p.RandomDelay,
		Debugger:         p.Debugger,
		Logger:           p.Logger,
		ArticleProcessor: p.ArticleProcessor,
		ContentProcessor: p.ContentProcessor,
		Source: config.Source{
			Name:         p.Source.Name,
			URL:          p.Source.URL,
			ArticleIndex: p.Source.ArticleIndex,
			Index:        p.Source.Index,
			RateLimit:    rateLimit,
			MaxDepth:     p.Source.MaxDepth,
			Time:         p.Source.Time,
			Selectors: config.SourceSelectors{
				Article: config.ArticleSelectors{
					Container:     p.Source.Selectors.Article.Container,
					Title:         p.Source.Selectors.Article.Title,
					Body:          p.Source.Selectors.Article.Body,
					Intro:         p.Source.Selectors.Article.Intro,
					Byline:        p.Source.Selectors.Article.Byline,
					PublishedTime: p.Source.Selectors.Article.PublishedTime,
					TimeAgo:       p.Source.Selectors.Article.TimeAgo,
					JSONLD:        p.Source.Selectors.Article.JSONLd,
					Section:       p.Source.Selectors.Article.Section,
					Keywords:      p.Source.Selectors.Article.Keywords,
					Description:   p.Source.Selectors.Article.Description,
					OGTitle:       p.Source.Selectors.Article.OgTitle,
					OGDescription: p.Source.Selectors.Article.OgDescription,
					OGImage:       p.Source.Selectors.Article.OgImage,
					OgURL:         p.Source.Selectors.Article.OgURL,
					Canonical:     p.Source.Selectors.Article.Canonical,
					WordCount:     p.Source.Selectors.Article.WordCount,
					PublishDate:   p.Source.Selectors.Article.PublishDate,
					Category:      p.Source.Selectors.Article.Category,
					Tags:          p.Source.Selectors.Article.Tags,
					Author:        p.Source.Selectors.Article.Author,
					BylineName:    p.Source.Selectors.Article.BylineName,
				},
			},
		},
	}, nil
}

// ValidateConfig validates the collector configuration.
// It ensures all required fields are present and valid.
//
// Returns:
//   - error: Any validation error that occurred
func (c *Config) ValidateConfig() error {
	// Validate base URL
	if c.BaseURL == "" {
		return errors.New("base URL is required")
	}
	// Validate max depth
	if c.MaxDepth < 0 {
		return errors.New("max depth must be non-negative")
	}
	// Validate parallelism
	if c.Parallelism < 1 {
		return errors.New("parallelism must be positive")
	}
	// Validate random delay
	if c.RandomDelay < 0 {
		return errors.New("random delay must be non-negative")
	}
	return nil
}
