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
			return setupMultiCrawlCmd(cmd, config, log)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeMultiCrawlCmd(cmd, log)
		},
	}

	return multiCrawlCmd
}

// setupMultiCrawlCmd handles the setup for the multi-crawl command
func setupMultiCrawlCmd(_ *cobra.Command, cfg *config.Config, log logger.Interface) error {
	// Load the sources from sources.yml
	sources, err := loadSourcesFromYAML("sources.yml")
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Debugging: Print loaded sources
	for _, source := range sources {
		log.Debug("Loaded source", "name", source.Name, "url", source.URL, "index", source.Index)
	}

	// Update the Config with the first source's BaseURL
	if len(sources) > 0 {
		// Use the Config methods to set the configuration values
		cfg.Crawler.SetBaseURL(sources[0].URL)     // Set the BaseURL using the Config method
		cfg.Crawler.SetIndexName(sources[0].Index) // Set the IndexName using the Config method

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

// loadSourcesFromYAML reads the sources from a YAML file
func loadSourcesFromYAML(filePath string) ([]Source, error) {
	var sources struct {
		Sources []Source `yaml:"sources"` // Wrap in a struct to match the YAML structure
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, err
	}

	return sources.Sources, nil // Return the slice of sources
}

// Source struct to represent the structure of your sources.yml
type Source struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Index string `yaml:"index"`
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
