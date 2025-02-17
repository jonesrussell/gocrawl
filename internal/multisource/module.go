package multisource

import (
	"context"
	"os"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

// SourceConfig represents the configuration for a single source
type SourceConfig struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Index string `yaml:"index"`
}

// MultiSource manages multiple sources for crawling
type MultiSource struct {
	Sources []SourceConfig   `yaml:"sources"`
	Crawler *crawler.Crawler `yaml:"-"`
	Logger  logger.Interface `yaml:"-"`
}

// NewMultiSource creates a new MultiSource instance
func NewMultiSource(logger logger.Interface, crawler *crawler.Crawler, configPath string) (*MultiSource, error) {
	if configPath == "" {
		configPath = "sources.yml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read sources.yml")
	}

	var ms MultiSource
	if unmarshalErr := yaml.Unmarshal(data, &ms); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failed to unmarshal sources.yml")
	}

	ms.Crawler = crawler
	ms.Logger = logger
	return &ms, nil
}

// Start begins the crawling process for all sources
func (ms *MultiSource) Start(ctx context.Context) error {
	for _, source := range ms.Sources {
		ms.Logger.Info("Starting crawl", "source", source.Name)

		// Use the setter methods to set the BaseURL and IndexName for the Crawler
		ms.Crawler.Config.Crawler.SetBaseURL(source.URL)
		ms.Crawler.Config.Crawler.SetIndexName(source.Index)

		// Start the Crawler
		if err := ms.Crawler.Start(ctx); err != nil {
			return errors.Wrapf(err, "error crawling source %s", source.Name)
		}
		ms.Logger.Info("Finished crawl", "source", source.Name)
	}
	return nil
}

// Stop halts the crawling process
func (ms *MultiSource) Stop() {
	ms.Crawler.Stop()
}

// Module provides the fx module for MultiSource
var Module = fx.Module(
	"multisource",
	fx.Provide(NewMultiSource),
)
