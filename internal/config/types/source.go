package types

import (
	"errors"
)

// Source represents a source to be crawled.
type Source struct {
	// Name is the unique identifier for the source
	Name string `yaml:"name"`
	// URL is the base URL for the source
	URL string `yaml:"url"`
	// AllowedDomains specifies which domains are allowed to be crawled
	AllowedDomains []string `yaml:"allowed_domains"`
	// StartURLs are the initial URLs to start crawling from
	StartURLs []string `yaml:"start_urls"`
	// RateLimit defines the delay between requests for this source
	RateLimit string `yaml:"rate_limit"`
	// MaxDepth defines how many levels deep to crawl for this source
	MaxDepth int `yaml:"max_depth"`
	// Time holds time-related configuration
	Time []string `yaml:"time"`
	// Index is the name of the index for content
	Index string `yaml:"index"`
	// ArticleIndex is the name of the index for articles
	ArticleIndex string `yaml:"article_index"`
	// PageIndex is the name of the index for pages
	PageIndex string `yaml:"page_index"`
	// Selectors define CSS selectors for extracting content
	Selectors SourceSelectors `yaml:"selectors"`
	// Rules define crawling rules for this source
	Rules Rules `yaml:"rules"`
}

// Validate validates the source configuration.
func (s *Source) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.URL == "" {
		return errors.New("url is required")
	}
	if len(s.StartURLs) == 0 {
		return errors.New("at least one start_url is required")
	}
	if s.MaxDepth < 0 {
		return errors.New("max_depth must be non-negative")
	}
	if s.RateLimit == "" {
		return errors.New("rate_limit is required")
	}
	if err := s.Selectors.Validate(); err != nil {
		return err
	}
	return s.Rules.Validate()
}
