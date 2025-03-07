package sources

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"gopkg.in/yaml.v3"
)

// IndexManager defines the interface for index management
type IndexManager interface {
	EnsureIndex(ctx context.Context, indexName string) error
}

// CrawlerInterface defines the interface for crawler operations
type CrawlerInterface interface {
	Start(ctx context.Context, url string) error
	Stop()
}

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
			JSONLd        string `yaml:"json_ld,omitempty"`
			Section       string `yaml:"section,omitempty"`
			Keywords      string `yaml:"keywords,omitempty"`
			Description   string `yaml:"description,omitempty"`
			OgTitle       string `yaml:"og_title,omitempty"`
			OgDescription string `yaml:"og_description,omitempty"`
			OgImage       string `yaml:"og_image,omitempty"`
			OgURL         string `yaml:"og_url,omitempty"`
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
	Sources  []Config         `yaml:"sources"`
	Crawler  CrawlerInterface `yaml:"-"`
	Logger   logger.Interface `yaml:"-"`
	IndexMgr IndexManager     `yaml:"-"`
}

// Load loads sources from a YAML file
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

// FindByName finds a source by its name
func (s *Sources) FindByName(name string) (*Config, error) {
	for _, source := range s.Sources {
		if source.Name == name {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("no source found with name: %s", name)
}

// SetCrawler sets the crawler instance
func (s *Sources) SetCrawler(c CrawlerInterface) {
	s.Crawler = c
}

// SetIndexManager sets the index manager
func (s *Sources) SetIndexManager(m IndexManager) {
	s.IndexMgr = m
}

// Start starts crawling a source
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
		if err := s.Crawler.Start(ctx, source.URL); err != nil {
			if !errors.Is(err, context.Canceled) {
				s.Logger.Error("Failed to start crawler", "error", err)
				crawlErr = err
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

// Stop stops the crawler
func (s *Sources) Stop() {
	if s.Crawler != nil {
		if s.Logger != nil {
			s.Logger.Debug("Stopping source crawler")
		}
		s.Crawler.Stop()
	}
}
