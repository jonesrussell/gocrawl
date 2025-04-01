// Package loader provides functionality for loading source configurations from files.
package loader

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents a source configuration loaded from a file.
type Config struct {
	Name         string          `yaml:"name"`
	URL          string          `yaml:"url"`
	RateLimit    string          `yaml:"rate_limit"`
	MaxDepth     int             `yaml:"max_depth"`
	Time         []string        `yaml:"time"`
	ArticleIndex string          `yaml:"article_index"`
	Index        string          `yaml:"index"`
	Selectors    SourceSelectors `yaml:"selectors"`
}

// SourceSelectors defines the selectors for a source.
type SourceSelectors struct {
	Article ArticleSelectors `yaml:"article"`
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

// FileConfig represents the configuration file structure.
type FileConfig struct {
	Sources []Config `yaml:"sources"`
}

// LoadFromFile loads source configurations from a YAML file.
func LoadFromFile(path string) ([]Config, error) {
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, fmt.Errorf("failed to open file: %w", readErr)
	}

	var fileConfig FileConfig
	if unmarshalErr := yaml.Unmarshal(data, &fileConfig); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", unmarshalErr)
	}

	// Validate each config
	for i, cfg := range fileConfig.Sources {
		if validateErr := ValidateConfig(&cfg); validateErr != nil {
			return nil, fmt.Errorf("invalid config at index %d: %w", i, validateErr)
		}
	}

	return fileConfig.Sources, nil
}

// ValidateConfig validates a source configuration.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	if cfg.Name == "" {
		return errors.New("name is required")
	}

	if cfg.URL == "" {
		return errors.New("URL is required")
	}

	// Parse URL and validate scheme
	u, parseErr := url.Parse(cfg.URL)
	if parseErr != nil || u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("invalid URL: must be a valid HTTP(S) URL")
	}

	if cfg.RateLimit == "" {
		return errors.New("rate limit is required")
	}

	if cfg.MaxDepth <= 0 {
		return errors.New("max depth must be positive")
	}

	// Validate time format if provided
	if len(cfg.Time) > 0 {
		for _, t := range cfg.Time {
			if _, timeErr := time.Parse("15:04", t); timeErr != nil {
				return fmt.Errorf("invalid time format: %w", timeErr)
			}
		}
	}

	return nil
}
