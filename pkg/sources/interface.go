// Package sources provides source management functionality for the application.
package sources

import (
	"time"
)

// Config represents a source configuration.
type Config struct {
	Name         string            `yaml:"name"`
	URL          string            `yaml:"url"`
	RateLimit    time.Duration     `yaml:"-"`
	MaxDepth     int               `yaml:"max_depth"`
	ArticleIndex string            `yaml:"article_index"`
	Index        string            `yaml:"index"`
	Time         []string          `yaml:"time,omitempty"`
	Selectors    SelectorConfig    `yaml:"selectors"`
	Metadata     map[string]string `yaml:"metadata,omitempty"`
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

// Interface defines the source management operations
type Interface interface {
	// FindByName finds a source configuration by name
	FindByName(name string) (*Config, error)

	// GetSources returns all source configurations
	GetSources() []Config

	// Validate validates a source configuration
	Validate(source *Config) error

	// SetSources sets the source configurations
	SetSources(sources []Config)
}
