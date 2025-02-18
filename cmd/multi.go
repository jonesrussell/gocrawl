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

// NewMultiCrawlCmd creates a new command for multi-source crawling
func NewMultiCrawlCmd(log logger.Interface, cfg *config.Config, multiSource *multisource.MultiSource, c *crawler.Crawler) *cobra.Command {
	var sourceName string

	var multiCrawlCmd = &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			log.Debug("Starting multi-crawl command...", "sourceName", sourceName)
			return executeMultiCrawlCmd(cmd, log, multiSource, sourceName)
		},
	}

	multiCrawlCmd.Flags().StringVar(&sourceName, "source", "", "Specify the source to crawl")
	if err := multiCrawlCmd.MarkFlagRequired("source"); err != nil {
		log.Error("Error marking source flag as required", "error", err)
	}

	return multiCrawlCmd
}

// executeMultiCrawlCmd executes the multi-source crawl command
func executeMultiCrawlCmd(cmd *cobra.Command, log logger.Interface, multiSource *multisource.MultiSource, sourceName string) error {
	app := newMultiCrawlFxApp(log, "sources.yml", sourceName, multiSource.Crawler)
	ctx := cmd.Context()

	defer func() {
		if err := app.Stop(ctx); err != nil {
			log.Error("Error stopping application", "context", ctx, "error", err)
		}
	}()

	if err := app.Start(ctx); err != nil {
		log.Error("Error starting application", "context", ctx, "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}

	if err := multiSource.Start(ctx, sourceName); err != nil {
		log.Error("Error starting multi-source crawl", "context", ctx, "sourceName", sourceName, "error", err)
		return fmt.Errorf("error starting multi-source crawl: %w", err)
	}

	return nil
}

// filterSources filters the sources based on source name
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

// newMultiCrawlFxApp initializes a new Fx application for multi-source crawling
func newMultiCrawlFxApp(log logger.Interface, configPath string, sourceName string, c *crawler.Crawler) *fx.App {
	log.Debug("Initializing multi-crawl application...")

	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		crawler.Module,
		multisource.Module,
		fx.Provide(func() *multisource.MultiSource {
			ms, err := multisource.NewMultiSource(log, c, configPath, sourceName)
			if err != nil {
				log.Error("Error creating MultiSource", "error", err)
				return nil
			}
			return ms
		}),
		fx.Provide(func() string {
			return sourceName
		}),
		fx.Invoke(setupMultiLifecycleHooks),
	)
}

// setupMultiLifecycleHooks sets up lifecycle hooks for the multi-crawl application
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
			deps.Logger.Debug("Starting multi-source crawl", "sourceName", deps.SourceName)
			return deps.MultiSource.Start(ctx, deps.SourceName)
		},
		OnStop: func(ctx context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.MultiSource.Stop()
			return nil
		},
	})
}
