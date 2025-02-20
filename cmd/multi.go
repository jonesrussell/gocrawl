package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var sourceName string

var multiCmd = &cobra.Command{
	Use:   "multi",
	Short: "Crawl multiple sources defined in sources.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		globalLogger.Debug("Starting multi-crawl command...", "sourceName", sourceName)

		// Create an Fx application
		app := fx.New(
			config.Module,
			logger.Module,
			storage.Module,
			crawler.Module,
			multisource.Module,
			fx.Provide(func(c *crawler.Crawler) *multisource.MultiSource {
				ms, err := multisource.NewMultiSource(globalLogger, c, "sources.yml")
				if err != nil {
					return nil
				}
				return ms
			}),
			fx.Invoke(func(ms *multisource.MultiSource, c *crawler.Crawler) error {
				if c == nil {
					return fmt.Errorf("Crawler is not initialized")
				}

				// Filter sources based on sourceName
				filteredSources, err := filterSources(ms.Sources, sourceName)
				if err != nil {
					return err
				}

				// Set the base URL from the filtered source
				globalConfig.Crawler.SetBaseURL(filteredSources[0].URL)

				// Create the collector using the collector module
				collectorResult, err := collector.New(collector.Params{
					BaseURL:   globalConfig.Crawler.BaseURL,
					MaxDepth:  globalConfig.Crawler.MaxDepth,
					RateLimit: globalConfig.Crawler.RateLimit,
					Debugger:  logger.NewCollyDebugger(globalLogger),
					Logger:    globalLogger,
				})
				if err != nil {
					return fmt.Errorf("error creating collector: %w", err)
				}

				// Set the collector in the crawler instance
				c.SetCollector(collectorResult.Collector)

				// Start the multi-source crawl
				return ms.Start(ctx, sourceName)
			}),
		)

		// Start the application
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		defer func() {
			if err := app.Stop(ctx); err != nil {
				globalLogger.Error("Error stopping application", "context", ctx, "error", err)
			}
		}()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(multiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// multiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// multiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	multiCmd.Flags().StringVar(&sourceName, "source", "", "Specify the source to crawl")
	if err := multiCmd.MarkFlagRequired("source"); err != nil {
		globalLogger.Error("Error marking source flag as required", "error", err)
	}
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
			ms, err := multisource.NewMultiSource(log, c, configPath)
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
