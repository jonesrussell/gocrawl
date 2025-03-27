// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources/loader"
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

// Sources manages web content source configurations.
type Sources struct {
	sources []Config
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
	loaderConfigs, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	var configs []Config
	for _, src := range loaderConfigs {
		rateLimit, err := time.ParseDuration(src.RateLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rate limit for source %s: %w", src.Name, err)
		}

		configs = append(configs, Config{
			Name:         src.Name,
			URL:          src.URL,
			RateLimit:    rateLimit,
			MaxDepth:     src.MaxDepth,
			ArticleIndex: src.ArticleIndex,
			Index:        src.Index,
			Time:         src.Time,
			Selectors: SelectorConfig{
				Article: ArticleSelectors{
					Container:     src.Selectors.Article.Container,
					Title:         src.Selectors.Article.Title,
					Body:          src.Selectors.Article.Body,
					Intro:         src.Selectors.Article.Intro,
					Byline:        src.Selectors.Article.Byline,
					PublishedTime: src.Selectors.Article.PublishedTime,
					TimeAgo:       src.Selectors.Article.TimeAgo,
					JSONLD:        src.Selectors.Article.JSONLD,
					Section:       src.Selectors.Article.Section,
					Keywords:      src.Selectors.Article.Keywords,
					Description:   src.Selectors.Article.Description,
					OGTitle:       src.Selectors.Article.OGTitle,
					OGDescription: src.Selectors.Article.OGDescription,
					OGImage:       src.Selectors.Article.OGImage,
					OgURL:         src.Selectors.Article.OgURL,
					Canonical:     src.Selectors.Article.Canonical,
					WordCount:     src.Selectors.Article.WordCount,
					PublishDate:   src.Selectors.Article.PublishDate,
					Category:      src.Selectors.Article.Category,
					Tags:          src.Selectors.Article.Tags,
					Author:        src.Selectors.Article.Author,
					BylineName:    src.Selectors.Article.BylineName,
				},
			},
		})
	}

	return &Sources{sources: configs}, nil
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
	for _, source := range s.sources {
		if source.Name == name {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("no source found with name: %s", name)
}

// GetSources returns all available sources.
func (s *Sources) GetSources() []Config {
	return s.sources
}

// Validate checks if a source configuration is valid.
// It ensures all required fields are present and properly formatted.
//
// Parameters:
//   - source: The source configuration to validate
//
// Returns:
//   - error: Any validation errors found
func (s *Sources) Validate(source *Config) error {
	if source == nil {
		return errors.New("source configuration is nil")
	}

	if source.Name == "" {
		return errors.New("source name is required")
	}

	if source.URL == "" {
		return errors.New("source URL is required")
	}

	if source.RateLimit == 0 {
		return errors.New("rate limit is required")
	}

	if source.MaxDepth <= 0 {
		return errors.New("max depth must be greater than 0")
	}

	return nil
}

// SetSources sets the sources list. This is primarily used for testing.
func (s *Sources) SetSources(sources []Config) {
	s.sources = sources
}
