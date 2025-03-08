// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"gopkg.in/yaml.v3"
)

// Config represents a source configuration.
type Config struct {
	Name      string            `yaml:"name"`
	URL       string            `yaml:"url"`
	RateLimit string            `yaml:"rate_limit"`
	MaxDepth  int               `yaml:"max_depth"`
	Selectors SelectorConfig    `yaml:"selectors"`
	Metadata  map[string]string `yaml:"metadata,omitempty"`
	// Time specifies the scheduled times for crawling in 24-hour format (HH:MM)
	Time []string `yaml:"time,omitempty"`
	// ArticleIndex specifies the Elasticsearch index name for articles from this source
	ArticleIndex string `yaml:"article_index,omitempty"`
	// Index specifies the Elasticsearch index name for general content from this source
	Index string `yaml:"index,omitempty"`
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
	JSONLd        string `yaml:"jsonld"`
	Section       string `yaml:"section"`
	Keywords      string `yaml:"keywords"`
	Description   string `yaml:"description"`
	OgTitle       string `yaml:"og_title"`
	OgDescription string `yaml:"og_description"`
	OgImage       string `yaml:"og_image"`
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
	Title       string           `yaml:"title"`
	Description string           `yaml:"description"`
	Content     string           `yaml:"content"`
	Article     ArticleSelectors `yaml:"article"`
}

// Sources manages web content source configurations.
type Sources struct {
	Sources      []Config         `yaml:"sources"`
	Logger       logger.Interface `yaml:"-"`
	crawler      crawler.Interface
	indexManager IndexManager
}

// LoadFromFile loads source configurations from a YAML file.
// It reads and parses the file, populating the Sources struct with the configuration data.
//
// Parameters:
//   - path: Path to the YAML configuration file
//
// Returns:
//   - *Sources: The loaded sources configuration
//   - error: Any error that occurred during loading
func LoadFromFile(path string) (*Sources, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources file: %w", err)
	}

	var sources Sources
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sources: %w", err)
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

// SetLogger sets the logger instance.
// It assigns the provided logger interface to the sources instance.
//
// Parameters:
//   - l: The logger interface to set
func (s *Sources) SetLogger(l logger.Interface) {
	s.Logger = l
}

// SetCrawler sets the crawler instance.
func (s *Sources) SetCrawler(c crawler.Interface) {
	s.crawler = c
}

// SetIndexManager sets the index manager instance.
func (s *Sources) SetIndexManager(im IndexManager) {
	s.indexManager = im
}

// Start begins crawling the specified source.
func (s *Sources) Start(ctx context.Context, sourceName string) error {
	source, err := s.FindByName(sourceName)
	if err != nil {
		return fmt.Errorf("failed to find source: %w", err)
	}
	return s.crawler.Start(ctx, source.URL)
}

// Stop gracefully stops all crawling operations.
func (s *Sources) Stop(ctx context.Context) error {
	if s.crawler != nil {
		return s.crawler.Stop(ctx)
	}
	return nil
}
