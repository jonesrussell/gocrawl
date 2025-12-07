// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"context"
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	articlespkg "github.com/jonesrussell/gocrawl/internal/content/articles"
	pagepkg "github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Cmd represents the scheduler command.
var Cmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Start the scheduler",
	Long: `Start the scheduler to manage and execute scheduled crawling tasks.
The scheduler will run continuously until interrupted with Ctrl+C.`,
	RunE: runScheduler,
}

// runScheduler executes the scheduler command
func runScheduler(cmd *cobra.Command, _ []string) error {
	// Get dependencies from context using helper
	log, cfg, err := cmdcommon.GetDependencies(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Create source manager
	sourceManager, err := sources.LoadSources(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Construct dependencies directly without FX
	storageClient, err := createStorageClientForScheduler(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: storageClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create crawler (simplified - scheduler needs crawler for jobs)
	crawlerInstance, err := createCrawlerForScheduler(cfg, log, sourceManager, storageResult)
	if err != nil {
		return fmt.Errorf("failed to create crawler: %w", err)
	}

	// Create processor factory
	processorFactory := crawler.NewProcessorFactory(log, storageResult.Storage, "content")

	// Create done channel
	done := make(chan struct{})

	// Create scheduler service directly
	schedulerService := NewSchedulerService(
		log,
		sourceManager,
		crawlerInstance,
		done,
		cfg,
		storageResult.Storage,
		processorFactory,
	)

	// Start the scheduler service
	log.Info("Starting scheduler service")
	if startSchedulerErr := schedulerService.Start(cmd.Context()); startSchedulerErr != nil {
		log.Error("Failed to start scheduler service", "error", startSchedulerErr)
		return fmt.Errorf("failed to start scheduler service: %w", startSchedulerErr)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

	// Graceful shutdown with timeout
	log.Info("Shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cmdcommon.DefaultShutdownTimeout)
	defer cancel()

	// Stop the scheduler service
	if stopSchedulerErr := schedulerService.Stop(shutdownCtx); stopSchedulerErr != nil {
		log.Error("Failed to stop scheduler service", "error", stopSchedulerErr)
		return fmt.Errorf("failed to stop scheduler service: %w", stopSchedulerErr)
	}

	log.Info("Scheduler stopped successfully")
	return nil
}

// createStorageClientForScheduler creates an Elasticsearch client for the scheduler command
func createStorageClientForScheduler(cfg config.Interface, log logger.Interface) (*es.Client, error) {
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: cfg,
		Logger: log,
	})
	if err != nil {
		return nil, err
	}
	return clientResult.Client, nil
}

// createCrawlerForScheduler creates a crawler instance for the scheduler.
// The scheduler uses a single crawler instance and calls Start with different source URLs.
func createCrawlerForScheduler(
	cfg config.Interface,
	log logger.Interface,
	sourceManager sources.Interface,
	storageResult storage.StorageResult,
) (crawler.Interface, error) {
	// Create event bus
	bus := events.NewEventBus(log)

	// Get crawler config
	crawlerCfg := cfg.GetCrawlerConfig()
	if crawlerCfg == nil {
		return nil, errors.New("crawler configuration is required")
	}

	// Use default index names for scheduler (it will use source-specific indices when crawling)
	articleIndexName := "articles"
	pageIndexName := "pages"

	// Construct article and page services
	articleService := articlespkg.NewContentService(log, storageResult.Storage, articleIndexName)
	pageService := pagepkg.NewContentService(log, storageResult.Storage, pageIndexName)

	// Use the crawler's ProvideCrawler to construct the crawler
	crawlerResult, err := crawler.ProvideCrawler(struct {
		fx.In
		Logger         logger.Interface
		Bus            *events.EventBus
		IndexManager   types.IndexManager
		Sources        sources.Interface
		Config         *crawlerconfig.Config
		ArticleService articlespkg.Interface
		PageService    pagepkg.Interface
		Storage        types.Interface
	}{
		Logger:         log,
		Bus:            bus,
		IndexManager:   storageResult.IndexManager,
		Sources:        sourceManager,
		Config:         crawlerCfg,
		ArticleService: articleService,
		PageService:    pageService,
		Storage:        storageResult.Storage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create crawler: %w", err)
	}

	return crawlerResult.Crawler, nil
}

// Command returns the scheduler command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
