package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"gopkg.in/yaml.v2"
)

// Config represents a source configuration
type Config struct {
	Name         string   `yaml:"name"`
	URL          string   `yaml:"url"`
	ArticleIndex string   `yaml:"article_index"` // Index for articles
	Index        string   `yaml:"index"`         // Index for non-article content
	RateLimit    string   `yaml:"rate_limit"`
	MaxDepth     int      `yaml:"max_depth"`
	Time         []string `yaml:"time"`
	Selectors    struct {
		Article struct {
			Container     string `yaml:"container,omitempty"`
			Title         string `yaml:"title"`
			Body          string `yaml:"body"`
			Intro         string `yaml:"intro,omitempty"`
			Byline        string `yaml:"byline,omitempty"`
			PublishedTime string `yaml:"published_time"`
			TimeAgo       string `yaml:"time_ago,omitempty"`
			JsonLd        string `yaml:"json_ld,omitempty"`
			Section       string `yaml:"section,omitempty"`
			Keywords      string `yaml:"keywords,omitempty"`
			Description   string `yaml:"description,omitempty"`
			OgTitle       string `yaml:"og_title,omitempty"`
			OgDescription string `yaml:"og_description,omitempty"`
			OgImage       string `yaml:"og_image,omitempty"`
			OgUrl         string `yaml:"og_url,omitempty"`
			Canonical     string `yaml:"canonical,omitempty"`
			WordCount     string `yaml:"word_count,omitempty"`
			PublishDate   string `yaml:"publish_date,omitempty"`
			Category      string `yaml:"category,omitempty"`
			Tags          string `yaml:"tags,omitempty"`
			Author        string `yaml:"author,omitempty"`
			BylineName    string `yaml:"byline_name,omitempty"`
		} `yaml:"article"`
	} `yaml:"selectors"`
}

// Sources represents the root YAML structure and handles crawling
type Sources struct {
	Sources []Config          `yaml:"sources"`
	Crawler crawler.Interface `yaml:"-"`
	Logger  logger.Interface  `yaml:"-"`
}

// Load loads sources from a YAML file
func Load(filename string) (*Sources, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var sources Sources
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, err
	}

	return &sources, nil
}

// FindByName finds a source by its name
func (s *Sources) FindByName(name string) (*Config, error) {
	for _, source := range s.Sources {
		if source.Name == name {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("no source found with name: %s", name)
}

// Start starts crawling for the specified source
func (s *Sources) Start(ctx context.Context, sourceName string) error {
	if s.Crawler == nil {
		return fmt.Errorf("crawler is not initialized")
	}

	source, err := s.FindByName(sourceName)
	if err != nil {
		return err
	}

	s.Logger.Info("Starting crawl", "source", source.Name)
	if err := s.Crawler.Start(ctx, source.URL); err != nil {
		return fmt.Errorf("error crawling source %s: %w", source.Name, err)
	}
	s.Logger.Info("Finished crawl", "source", source.Name)

	return nil
}

// Stop stops the crawler
func (s *Sources) Stop() {
	if s.Crawler != nil {
		s.Crawler.Stop()
	}
}
