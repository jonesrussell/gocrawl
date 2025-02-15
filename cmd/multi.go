package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
)

// NewMultiCrawlCmd creates a new command for crawling multiple sources
func NewMultiCrawlCmd(log logger.Interface, config *config.Config) *cobra.Command {
	var multiCrawlCmd = &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupMultiCrawlCmd(cmd, config)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeMultiCrawlCmd(cmd, log)
		},
	}

	return multiCrawlCmd
}

// setupMultiCrawlCmd handles the setup for the multi-crawl command
func setupMultiCrawlCmd(_ *cobra.Command, config *config.Config) error {
	// Load the sources from sources.yml
	sources, err := loadSourcesFromYAML("sources.yml")
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Update the Config with the first source's BaseURL
	if len(sources) > 0 {
		config.Crawler.BaseURL = sources[0].BaseURL // Update the Config with the first source's URL
	}

	return nil
}

// loadSourcesFromYAML reads the sources from a YAML file
func loadSourcesFromYAML(filePath string) ([]Source, error) {
	var sources []Source
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, err
	}

	return sources, nil
}

// Source struct to represent the structure of your sources.yml
type Source struct {
	BaseURL string `yaml:"base_url"` // Adjust the field name based on your YAML structure
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
		storage.Module,
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
			return deps.Crawler.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.Crawler.Stop()
			deps.MultiSource.Stop()
			return nil
		},
	})
}
