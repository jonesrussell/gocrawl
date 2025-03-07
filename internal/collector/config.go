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

// Error messages
const (
	errEmptyBaseURL       = "base URL cannot be empty"
	errMissingArticleProc = "article processor is required"
	errMissingLogger      = "logger is required"
)

// Params holds the parameters for creating a Collector
type Params struct {
	fx.In

	ArticleProcessor models.ContentProcessor
	ContentProcessor models.ContentProcessor
	BaseURL          string
	Context          context.Context
	Debugger         *logger.CollyDebugger
	Logger           logger.Interface
	MaxDepth         int
	Parallelism      int
	RandomDelay      time.Duration
	RateLimit        time.Duration
	Source           *sources.Config
}

// Result holds the collector instance and completion channel
type Result struct {
	fx.Out

	Collector *colly.Collector
	Done      chan struct{} // Channel to signal when crawling is complete
}

// ValidateParams validates the collector parameters
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

// CollectorConfig holds all configuration for the collector
type CollectorConfig struct {
	BaseURL          string
	MaxDepth         int
	RateLimit        string
	Parallelism      int
	RandomDelay      time.Duration
	Debugger         *logger.CollyDebugger
	Logger           logger.Interface
	Source           config.Source
	ArticleProcessor models.ContentProcessor
	ContentProcessor models.ContentProcessor
}

// NewCollectorConfig creates a new collector configuration
func NewCollectorConfig(p Params) (*CollectorConfig, error) {
	rateLimit, err := time.ParseDuration(p.Source.RateLimit)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit: %w", err)
	}

	return &CollectorConfig{
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

// ValidateConfig validates the collector configuration
func (c *CollectorConfig) ValidateConfig() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if c.MaxDepth < 0 {
		return fmt.Errorf("max depth must be non-negative")
	}
	if c.Parallelism < 1 {
		return fmt.Errorf("parallelism must be positive")
	}
	if c.RandomDelay < 0 {
		return fmt.Errorf("random delay must be non-negative")
	}
	return nil
}
