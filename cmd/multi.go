package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// NewMultiCrawlCmd creates a new command for crawling multiple sources
func NewMultiCrawlCmd(log logger.Interface, config *config.Config, multiSource *multisource.MultiSource) *cobra.Command {
	var multiCrawlCmd = &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupMultiCrawlCmd(cmd, config, log, multiSource)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeMultiCrawlCmd(cmd, log)
		},
	}

	return multiCrawlCmd
}

// setupMultiCrawlCmd handles the setup for the multi-crawl command
func setupMultiCrawlCmd(_ *cobra.Command, cfg *config.Config, log logger.Interface, multiSource *multisource.MultiSource) error {
	// Debugging: Print loaded sources
	for _, source := range multiSource.Sources {
		log.Debug("Loaded source", "name", source.Name, "url", source.URL, "index", source.Index)
	}

	// Update the Config with the first source's BaseURL
	if len(multiSource.Sources) > 0 {
		cfg.Crawler.SetBaseURL(multiSource.Sources[0].URL)     // Set the BaseURL using the Config method
		cfg.Crawler.SetIndexName(multiSource.Sources[0].Index) // Set the IndexName using the Config method

		// Log updated config directly from the Config struct
		log.Debug(
			"Updated config",
			"baseURL",
			cfg.Crawler.BaseURL,
			"indexName",
			cfg.Crawler.IndexName,
		) // Log updated config
	}

	return nil
}

// executeMultiCrawlCmd handles the execution of the multi-crawl command
func executeMultiCrawlCmd(cmd *cobra.Command, log logger.Interface) error {
	// Initialize fx container
	app := newMultiCrawlFxApp(log)

	// Start the application
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

// newMultiCrawlFxApp initializes the Fx application with dependencies for multi-crawl
func newMultiCrawlFxApp(log logger.Interface) *fx.App {
	log.Debug("Initializing multi-crawl application...") // Log initialization

	return fx.New(
		config.Module,
		logger.Module,
		crawler.Module,
		multisource.Module,
		fx.Invoke(setupMultiLifecycleHooks),
	)
}

// setupMultiLifecycleHooks sets up the lifecycle hooks for the Fx application
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
		OnStop: func(ctx context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.MultiSource.Stop()
			return nil
		},
	})
}
