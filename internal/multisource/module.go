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

func (ms *MultiSource) Start(ctx context.Context) error {
	for _, source := range ms.Sources {
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
