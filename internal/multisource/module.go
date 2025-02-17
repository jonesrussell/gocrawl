package multisource

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

type SourceConfig struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Index string `yaml:"index"`
}

type MultiSource struct {
	Sources []SourceConfig   `yaml:"sources"`
	Crawler *crawler.Crawler `yaml:"-"`
	Logger  logger.Interface `yaml:"-"`
}

func NewMultiSource(log logger.Interface, c *crawler.Crawler, configPath string) (*MultiSource, error) {
	log.Debug("NewMultiSource", "configPath", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read sources.yml")
	}

	var ms MultiSource
	if unmarshalErr := yaml.Unmarshal(data, &ms); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failed to unmarshal sources.yml")
	}

	ms.Crawler = c
	ms.Logger = log
	return &ms, nil
}

func (ms *MultiSource) filterSources(sourceName string) ([]SourceConfig, error) {
	var filteredSources []SourceConfig
	for _, source := range ms.Sources {
		if source.Name == sourceName {
			filteredSources = append(filteredSources, source)
		}
	}
	if len(filteredSources) == 0 {
		return nil, fmt.Errorf("no source found with name: %s", sourceName)
	}
	return filteredSources, nil
}

func (ms *MultiSource) Start(ctx context.Context, sourceName string) error {
	var sourcesToCrawl []SourceConfig

	// Filter sources based on the provided sourceName
	if sourceName != "" {
		filteredSources, err := ms.filterSources(sourceName)
		if err != nil {
			return err
		}
		sourcesToCrawl = filteredSources
	} else {
		sourcesToCrawl = ms.Sources
	}

	for _, source := range sourcesToCrawl {
		ms.Logger.Info("Starting crawl", "source", source.Name)

		ms.Crawler.Config.Crawler.SetBaseURL(source.URL)
		ms.Crawler.Config.Crawler.SetIndexName(source.Index)

		if err := ms.Crawler.Start(ctx); err != nil {
			return errors.Wrapf(err, "error crawling source %s", source.Name)
		}
		ms.Logger.Info("Finished crawl", "source", source.Name)
	}
	return nil
}

func (ms *MultiSource) Stop() {
	ms.Crawler.Stop()
}

var Module = fx.Module(
	"multisource",
	fx.Provide(
		func() string {
			return "sources.yml"
		},
		NewMultiSource,
	),
)
