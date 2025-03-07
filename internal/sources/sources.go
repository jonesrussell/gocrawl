// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading, validation, and crawling operations through a YAML-based configuration system.
package sources

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"gopkg.in/yaml.v3"
)

// IndexManager defines the interface for index management operations.
// It provides methods for ensuring that required indices exist in the storage system.
type IndexManager interface {
	// EnsureIndex ensures that an index exists in the storage system.
	// It creates the index if it doesn't exist and returns an error if the operation fails.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - indexName: Name of the index to ensure
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	EnsureIndex(ctx context.Context, indexName string) error
}

// CrawlerInterface defines the interface for crawler operations.
// It provides methods for controlling the crawling process.
type CrawlerInterface interface {
	// Start begins crawling from the specified URL.
	// It initiates the crawling process and returns an error if the operation fails.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - url: The URL to start crawling from
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	Start(ctx context.Context, url string) error

	// Stop stops the crawler.
	// It performs any necessary cleanup and stops the crawling process.
	Stop()
}

// Config represents a source configuration.
// It defines all the settings and selectors needed to crawl a specific source.
type Config struct {
	// Name is the unique identifier for the source
	Name string `yaml:"name"`
	// URL is the base URL to start crawling from
	URL string `yaml:"url"`
	// ArticleIndex is the name of the index for storing articles
	ArticleIndex string `yaml:"article_index"`
	// Index is the name of the index for storing non-article content
	Index string `yaml:"index"`
	// RateLimit is the time between requests as a string (e.g., "1s")
	RateLimit string `yaml:"rate_limit"`
	// MaxDepth is the maximum depth to crawl
	MaxDepth int `yaml:"max_depth"`
	// Time is a list of time-related fields to extract
	Time []string `yaml:"time"`
	// Selectors contains all the CSS selectors for content extraction
	Selectors struct {
		// Article contains selectors for article-specific content
		Article struct {
			// Container is the main article container selector
			Container string `yaml:"container,omitempty"`
			// Title is the selector for the article title
			Title string `yaml:"title"`
			// Body is the selector for the article body
			Body string `yaml:"body"`
			// Intro is the selector for the article introduction
			Intro string `yaml:"intro,omitempty"`
			// Byline is the selector for the article byline
			Byline string `yaml:"byline,omitempty"`
			// PublishedTime is the selector for the publication time
			PublishedTime string `yaml:"published_time"`
			// TimeAgo is the selector for relative time information
			TimeAgo string `yaml:"time_ago,omitempty"`
			// JSONLd is the selector for JSON-LD structured data
			JSONLd string `yaml:"json_ld,omitempty"`
			// Section is the selector for the article section
			Section string `yaml:"section,omitempty"`
			// Keywords is the selector for article keywords
			Keywords string `yaml:"keywords,omitempty"`
			// Description is the selector for the article description
			Description string `yaml:"description,omitempty"`
			// OgTitle is the selector for OpenGraph title
			OgTitle string `yaml:"og_title,omitempty"`
			// OgDescription is the selector for OpenGraph description
			OgDescription string `yaml:"og_description,omitempty"`
			// OgImage is the selector for OpenGraph image
			OgImage string `yaml:"og_image,omitempty"`
			// OgURL is the selector for OpenGraph URL
			OgURL string `yaml:"og_url,omitempty"`
			// Canonical is the selector for canonical URL
			Canonical string `yaml:"canonical,omitempty"`
			// WordCount is the selector for word count
			WordCount string `yaml:"word_count,omitempty"`
			// PublishDate is the selector for publication date
			PublishDate string `yaml:"publish_date,omitempty"`
			// Category is the selector for article category
			Category string `yaml:"category,omitempty"`
			// Tags is the selector for article tags
			Tags string `yaml:"tags,omitempty"`
			// Author is the selector for article author
			Author string `yaml:"author,omitempty"`
			// BylineName is the selector for byline author name
			BylineName string `yaml:"byline_name,omitempty"`
		} `yaml:"article"`
	} `yaml:"selectors"`
}

// Sources represents the root YAML structure and handles crawling operations.
// It manages the collection of source configurations and coordinates with the crawler.
type Sources struct {
	// Sources is the list of source configurations
	Sources []Config `yaml:"sources"`
	// Crawler is the crawler interface for performing crawling operations
	Crawler CrawlerInterface `yaml:"-"`
	// Logger is the logger interface for logging operations
	Logger logger.Interface `yaml:"-"`
	// IndexMgr is the index manager for managing storage indices
	IndexMgr IndexManager `yaml:"-"`
}

// Load loads sources from a YAML file.
// It reads and parses the YAML configuration file into a Sources instance.
//
// Parameters:
//   - filename: Path to the YAML configuration file
//
// Returns:
//   - *Sources: The loaded sources configuration
//   - error: Any error that occurred during loading
func Load(filename string) (*Sources, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources file: %w", err)
	}

	var sources Sources
	if unmarshalErr := yaml.Unmarshal(data, &sources); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal sources: %w", unmarshalErr)
	}

	return &sources, nil
}

// FindByName finds a source by its name.
// It searches through the sources list for a matching name.
//
// Parameters:
//   - name: The name of the source to find
//
// Returns:
//   - *Config: The found source configuration
//   - error: Any error that occurred during the search
func (s *Sources) FindByName(name string) (*Config, error) {
	for _, source := range s.Sources {
		if source.Name == name {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("no source found with name: %s", name)
}

// SetCrawler sets the crawler instance.
// It assigns the provided crawler interface to the sources instance.
//
// Parameters:
//   - c: The crawler interface to set
func (s *Sources) SetCrawler(c CrawlerInterface) {
	s.Crawler = c
}

// SetIndexManager sets the index manager.
// It assigns the provided index manager to the sources instance.
//
// Parameters:
//   - m: The index manager to set
func (s *Sources) SetIndexManager(m IndexManager) {
	s.IndexMgr = m
}

// Start starts crawling a source.
// It initiates the crawling process for the specified source and handles completion.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sourceName: Name of the source to crawl
//
// Returns:
//   - error: Any error that occurred during crawling
func (s *Sources) Start(ctx context.Context, sourceName string) error {
	source, err := s.FindByName(sourceName)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	var crawlErr error

	// Start crawling in a goroutine
	go func() {
		defer close(done)
		if startErr := s.Crawler.Start(ctx, source.URL); startErr != nil {
			if !errors.Is(startErr, context.Canceled) {
				s.Logger.Error("Failed to start crawler", "error", startErr)
				crawlErr = startErr
			}
		}
		s.Logger.Debug("Source crawl finished", "source", sourceName)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return crawlErr
	}
}

// Stop stops the crawler.
// It performs cleanup and stops any ongoing crawling operations.
func (s *Sources) Stop() {
	if s.Crawler != nil {
		if s.Logger != nil {
			s.Logger.Debug("Stopping source crawler")
		}
		s.Crawler.Stop()
	}
}
