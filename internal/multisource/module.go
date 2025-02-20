package multisource

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

// SourceConfig represents the configuration for a source
type SourceConfig struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Index     string `yaml:"index"`
	RateLimit string `yaml:"rate_limit"`
	MaxDepth  int    `yaml:"max_depth"`
}

// MultiSource represents a multi-source configuration
type MultiSource struct {
	Sources []SourceConfig   `yaml:"sources"`
	Crawler *crawler.Crawler `yaml:"-"`
	Logger  logger.Interface `yaml:"-"`
}

// NewMultiSource creates a new MultiSource instance
func NewMultiSource(log logger.Interface, c *crawler.Crawler, configPath string) (*MultiSource, error) {
	log.Debug("NewMultiSource", "configPath", configPath)

	file, openErr := os.Open(configPath)
	if openErr != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configPath, openErr)
	}
	defer file.Close()

	var config struct {
		Sources []SourceConfig `yaml:"sources"`
	}

	if decodeErr := yaml.NewDecoder(file).Decode(&config); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", configPath, decodeErr)
	}

	ms := &MultiSource{Sources: config.Sources, Crawler: c, Logger: log}
	return ms, nil
}

// Start starts the multi-source crawling for the specified source name
func (ms *MultiSource) Start(ctx context.Context, sourceName string) error {
	if ms == nil {
		return errors.New("MultiSource is nil")
	}
	if ms.Crawler == nil {
		return errors.New("crawler is not initialized")
	}
	ms.Logger.Debug("Starting multi-source crawl", "sourceName", sourceName)

	filteredSources, filterErr := filterSources(ms.Sources, sourceName)
	if filterErr != nil {
		return filterErr
	}

	// Start crawling with filtered sources
	for _, source := range filteredSources {
		ms.Logger.Info("Starting crawl", "source", source.Name)

		if crawlErr := ms.Crawler.Start(ctx, source.URL); crawlErr != nil {
			return fmt.Errorf("error crawling source %s: %w", source.Name, crawlErr)
		}

		ms.Logger.Info("Finished crawl", "source", source.Name)
	}
	return nil
}

// filterSources filters the sources based on source name
func filterSources(sources []SourceConfig, sourceName string) ([]SourceConfig, error) {
	var filteredSources []SourceConfig

	for _, source := range sources {
		if source.Name == sourceName {
			filteredSources = append(filteredSources, source)
		}
	}

	if len(filteredSources) == 0 {
		return nil, fmt.Errorf("no source found with name: %s", sourceName)
	}

	return filteredSources, nil
}

func (ms *MultiSource) Stop() {
	ms.Crawler.Stop()
}

var Module = fx.Module(
	"multisource",
)
