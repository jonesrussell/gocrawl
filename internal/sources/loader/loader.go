// Package loader provides functionality for loading source configurations from files.
package loader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents a source configuration.
type Config struct {
	Name         string            `yaml:"name"`
	URL          string            `yaml:"url"`
	RateLimit    string            `yaml:"rate_limit"`
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

// LoadFromFile loads source configurations from a YAML file.
// It reads and parses the file, populating the Sources struct with the configuration data.
//
// Parameters:
//   - path: Path to the YAML configuration file
//
// Returns:
//   - []Config: The loaded source configurations
//   - error: Any error that occurred during loading
func LoadFromFile(path string) ([]Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources file: %w", err)
	}

	// First unmarshal into a temporary struct to handle the YAML structure
	var temp struct {
		Sources []Config `yaml:"sources"`
	}
	if unmarshalErr := yaml.Unmarshal(data, &temp); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal sources: %w", unmarshalErr)
	}

	return temp.Sources, nil
}
