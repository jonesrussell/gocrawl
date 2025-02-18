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

var sourceName string

var multiCmd = &cobra.Command{
	Use:   "multi",
	Short: "Crawl multiple sources defined in sources.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		globalLogger.Debug("Starting multi-crawl command...", "sourceName", sourceName)

		// Initialize Elasticsearch client
		elasticClient, err := storage.ProvideElasticsearchClient(globalConfig, globalLogger)
		if err != nil {
			return fmt.Errorf("error creating Elasticsearch client: %w", err)
		}

		// Initialize storage
		storageInstance, err := storage.NewStorage(elasticClient, globalLogger)
		if err != nil {
			return fmt.Errorf("error creating storage: %w", err)
		}

		// Initialize the debugger
		debuggerInstance := logger.NewCollyDebugger(globalLogger)

		// Initialize the crawler
		crawlerParams := crawler.Params{
			Logger:   globalLogger,
			Storage:  storageInstance,
			Debugger: debuggerInstance,
			Config:   globalConfig,
		}

		crawlerInstance, err := crawler.NewCrawler(crawlerParams)
		if err != nil {
			return fmt.Errorf("error creating Crawler: %w", err)
		}

		// Initialize MultiSource
		multiSource, err := multisource.NewMultiSource(globalLogger, crawlerInstance.Crawler, "sources.yml", sourceName)
		if err != nil {
			globalLogger.Error("Error creating MultiSource", "error", err)
			return fmt.Errorf("error creating MultiSource: %w", err)
		}

		app := newMultiCrawlFxApp(globalLogger, "sources.yml", sourceName, crawlerInstance.Crawler)

		defer func() {
			if err := app.Stop(ctx); err != nil {
				globalLogger.Error("Error stopping application", "context", ctx, "error", err)
			}
		}()

		if err := app.Start(ctx); err != nil {
			globalLogger.Error("Error starting application", "context", ctx, "error", err)
			return fmt.Errorf("error starting application: %w", err)
		}

		if err := multiSource.Start(ctx, sourceName); err != nil {
			globalLogger.Error("Error starting multi-source crawl", "context", ctx, "sourceName", sourceName, "error", err)
			return fmt.Errorf("error starting multi-source crawl: %w", err)
		}

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
