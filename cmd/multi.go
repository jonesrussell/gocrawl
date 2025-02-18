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

func executeMultiCrawlCmd(cmd *cobra.Command, log logger.Interface, multiSource *multisource.MultiSource, sourceName string) error {
	app := newMultiCrawlFxApp(log)

	if err := app.Start(cmd.Context()); err != nil {
		log.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}

	if err := multiSource.Start(cmd.Context(), sourceName); err != nil {
		log.Error("Error starting multi-source crawl", "error", err)
		return fmt.Errorf("error starting multi-source crawl: %w", err)
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
	SourceName  string
}) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Debug("Starting multi-crawl application...")
			return deps.MultiSource.Start(ctx, deps.SourceName)
		},
		OnStop: func(_ context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.MultiSource.Stop()
			return nil
		},
	})
}
