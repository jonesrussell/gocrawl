package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewMultiCrawlCmd(log logger.Interface, cfg *config.Config, multiSource *multisource.MultiSource) *cobra.Command {
	var sourceName string

	var multiCrawlCmd = &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupMultiCrawlCmd(cmd, cfg, log, multiSource)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeMultiCrawlCmd(cmd, log, multiSource, sourceName)
		},
	}

	multiCrawlCmd.Flags().StringVar(&sourceName, "source", "", "Specify the source to crawl")
	if err := multiCrawlCmd.MarkFlagRequired("source"); err != nil {
		log.Error("Error marking source flag as required", "error", err)
	}

	return multiCrawlCmd
}

func setupMultiCrawlCmd(
	_ *cobra.Command,
	cfg *config.Config,
	log logger.Interface,
	multiSource *multisource.MultiSource,
) error {
	for _, source := range multiSource.Sources {
		log.Debug("Loaded source", "name", source.Name, "url", source.URL, "index", source.Index)
	}

	if len(multiSource.Sources) > 0 {
		cfg.Crawler.SetBaseURL(multiSource.Sources[0].URL)
		cfg.Crawler.SetIndexName(multiSource.Sources[0].Index)

		log.Debug(
			"Updated config",
			"baseURL", cfg.Crawler.BaseURL,
			"indexName", cfg.Crawler.IndexName,
			"maxDepth", cfg.Crawler.MaxDepth,
		)
	}

	return nil
}

func executeMultiCrawlCmd(cmd *cobra.Command, log logger.Interface, multiSource *multisource.MultiSource, sourceName string) error {
	app := newMultiCrawlFxApp(log)

	if sourceName != "" {
		var err error
		multiSource.Sources, err = filterSources(multiSource.Sources, sourceName)
		if err != nil {
			return err
		}
	}

	if err := app.Start(cmd.Context()); err != nil {
		log.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}
	defer func() {
		if err := app.Stop(cmd.Context()); err != nil {
			log.Error("Error stopping application", "error", err)
		}
	}()

	return nil
}

func filterSources(sources []multisource.SourceConfig, sourceName string) ([]multisource.SourceConfig, error) {
	var filteredSources []multisource.SourceConfig
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

func newMultiCrawlFxApp(log logger.Interface) *fx.App {
	log.Debug("Initializing multi-crawl application...")

	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		crawler.Module,
		multisource.Module,
		fx.Invoke(setupMultiLifecycleHooks),
	)
}

func setupMultiLifecycleHooks(lc fx.Lifecycle, deps struct {
	fx.In
	Logger      logger.Interface
	Crawler     *crawler.Crawler
	MultiSource *multisource.MultiSource
	Config      *config.Config
}) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Debug("Starting multi-crawl application...")
			return deps.MultiSource.Start(ctx)
		},
		OnStop: func(_ context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.MultiSource.Stop()
			return nil
		},
	})
}
