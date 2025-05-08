// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"time"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/storage"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const (
	defaultParallelism = 2
	defaultRandomDelay = 5 * time.Second
)

// Cmd represents the crawl command.
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
	Args: cobra.ExactArgs(1),
	RunE: runCrawl,
}

// runCrawl executes the crawl command
func runCrawl(cmd *cobra.Command, args []string) error {
	// Get logger from context
	loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New("logger not found in context or invalid type")
	}

	// Get config from context
	configValue := cmd.Context().Value(cmdcommon.ConfigKey)
	cfg, ok := configValue.(config.Interface)
	if !ok {
		return errors.New("config not found in context or invalid type")
	}

	// Create source manager
	sourceManager, err := sources.LoadSources(cfg)
	if err != nil {
		if errors.Is(err, loader.ErrNoSources) {
			log.Info("No sources found in configuration. Please add sources to your config file.")
			log.Info("You can use the 'sources list' command to view configured sources.")
			return nil
		}
		return fmt.Errorf("failed to load sources: %w", err)
	}

	var jobService common.JobService

	// Create Fx app with the module
	fxApp := fx.New(
		Module,
		fx.Provide(
			func() logger.Interface { return log },
			func() config.Interface { return cfg },
			func() sources.Interface { return sourceManager },
			func() (storagetypes.Interface, error) {
				opts := storage.Options{
					Addresses: cfg.GetElasticsearchConfig().Addresses,
					Username:  cfg.GetElasticsearchConfig().Username,
					Password:  cfg.GetElasticsearchConfig().Password,
					APIKey:    cfg.GetElasticsearchConfig().APIKey,
				}
				client, err := storage.NewClient(opts)
				if err != nil {
					return nil, err
				}
				return storage.NewStorage(client.GetClient(), log, &opts), nil
			},
			fx.Annotate(
				func() string { return args[0] },
				fx.ResultTags(`name:"sourceName"`),
			),
		),
		fx.WithLogger(func() fxevent.Logger {
			return logger.NewFxLogger(log)
		}),
		fx.Invoke(func(js common.JobService) {
			jobService = js
		}),
	)

	// Start the application
	log.Info("Starting application")
	startErr := fxApp.Start(cmd.Context())
	if startErr != nil {
		log.Error("Failed to start application", "error", startErr)
		return fmt.Errorf("failed to start application: %w", startErr)
	}

	// Start the job service
	if startJobErr := jobService.Start(cmd.Context()); startJobErr != nil {
		log.Error("Failed to start job service", "error", startJobErr)
		return fmt.Errorf("failed to start job service: %w", startJobErr)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

	// Stop the job service
	if stopJobErr := jobService.Stop(cmd.Context()); stopJobErr != nil {
		log.Error("Failed to stop job service", "error", stopJobErr)
		return fmt.Errorf("failed to stop job service: %w", stopJobErr)
	}

	// Stop the application
	log.Info("Stopping application")
	stopErr := fxApp.Stop(cmd.Context())
	if stopErr != nil {
		log.Error("Failed to stop application", "error", stopErr)
		return fmt.Errorf("failed to stop application: %w", stopErr)
	}

	log.Info("Application stopped successfully")
	return nil
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}

// SetupCollector creates and configures a new collector instance.
func SetupCollector(
	ctx context.Context,
	logger logger.Interface,
	indexManager storagetypes.IndexManager,
	sources sources.Interface,
	eventBus *events.EventBus,
	articleProcessor content.Processor,
	contentProcessor content.Processor,
	cfg *crawlerconfig.Config,
) (crawler.Interface, error) {
	// Create crawler instance
	return crawler.NewCrawler(
		logger,
		eventBus,
		indexManager,
		sources,
		articleProcessor,
		contentProcessor,
		cfg,
	), nil
}
