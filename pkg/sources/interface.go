// Package sources provides source management functionality for the application.
package sources

import (
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/pkg/logger"
)

// Interface defines the interface for source management operations.
// It provides methods for managing and accessing source configurations.
type Interface interface {
	// GetSources returns all configured sources.
	GetSources() ([]Config, error)
	// FindByName finds a source by its name.
	FindByName(name string) (*Config, error)
	// Validate validates a source configuration.
	Validate(source *Config) error
}

// Config represents a source configuration.
type Config struct {
	// Name is the unique identifier for the source.
	Name string
	// URL is the base URL for the source.
	URL string
	// RateLimit is the rate limit for requests to this source.
	RateLimit time.Duration
	// MaxDepth is the maximum depth for crawling this source.
	MaxDepth int
	// Time is the list of times when this source should be crawled.
	Time []string
}

// Params holds the parameters for creating a sources instance.
type Params struct {
	// Config is the configuration for the sources instance.
	Config Interface
	// Logger is the logger for the sources instance.
	Logger logger.Interface
}

// ValidateParams validates the parameters for creating a sources instance.
func ValidateParams(p Params) error {
	if p.Config == nil {
		return errors.New("config is required")
	}
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors struct {
	Container     string `yaml:"container"`
	Title         string `yaml:"title"`
	Body          string `yaml:"body"`
	Intro         string `yaml:"intro"`
	Byline        string `yaml:"byline"`
	PublishedTime string `yaml:"published_time"`
	TimeAgo       string `yaml:"time_ago"`
	JSONLD        string `yaml:"jsonld"`
	Section       string `yaml:"section"`
	Keywords      string `yaml:"keywords"`
	Description   string `yaml:"description"`
	OGTitle       string `yaml:"og_title"`
	OGDescription string `yaml:"og_description"`
	OGImage       string `yaml:"og_image"`
	OgURL         string `yaml:"og_url"`
	Canonical     string `yaml:"canonical"`
	WordCount     string `yaml:"word_count"`
	PublishDate   string `yaml:"publish_date"`
	Category      string `yaml:"category"`
	Tags          string `yaml:"tags"`
	Author        string `yaml:"author"`
	BylineName    string `yaml:"byline_name"`
}

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig struct {
	Article ArticleSelectors `yaml:"article"`
}
