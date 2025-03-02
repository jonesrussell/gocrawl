package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"gopkg.in/yaml.v2"
)

// Config represents a source configuration
type Config struct {
	Name      string   `yaml:"name"`
	URL       string   `yaml:"url"`
	Index     string   `yaml:"index"`
	RateLimit string   `yaml:"rate_limit"`
	MaxDepth  int      `yaml:"max_depth"`
	Time      []string `yaml:"time"`
}

// Crawler interface defines the methods required for a crawler
type Crawler interface {
	Start(ctx context.Context, url string) error
	Stop()
}

// Sources represents the root YAML structure and handles crawling
type Sources struct {
	Sources []Config         `yaml:"sources"`
	Crawler Crawler          `yaml:"-"`
	Logger  logger.Interface `yaml:"-"`
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
