// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"context"
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/constants"
	articlespkg "github.com/jonesrussell/gocrawl/internal/content/articles"
	pagepkg "github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
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
	// Get dependencies
	deps, err := cmdcommon.NewCommandDeps()
	if err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Create source manager
	sourceManager, err := sources.LoadSources(deps.Config, deps.Logger)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Create storage using common function
	storageResult, err := cmdcommon.CreateStorage(deps.Config, deps.Logger)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create article and page services
	// Use default index names for scheduler (it will use source-specific indices when crawling)
	articleService := articlespkg.NewContentService(
		deps.Logger, storageResult.Storage, constants.DefaultArticleIndex)
	pageService := pagepkg.NewContentService(
		deps.Logger, storageResult.Storage, constants.DefaultPageIndex)

	// Create crawler
	crawlerInstance, err := createCrawlerInstance(
		deps.Logger, deps.Config, sourceManager, storageResult, articleService, pageService)
	if err != nil {
		return fmt.Errorf("failed to create crawler: %w", err)
	}

	// Create processor factory
	processorFactory := crawler.NewProcessorFactory(deps.Logger, storageResult.Storage, constants.DefaultContentIndex)

	// Create done channel
	done := make(chan struct{})

	// Create scheduler service directly
	schedulerService := NewSchedulerService(
		deps.Logger,
		sourceManager,
		crawlerInstance,
		done,
		deps.Config,
		storageResult.Storage,
		processorFactory,
	)

	// Start the scheduler service
	deps.Logger.Info("Starting scheduler service")
	if startSchedulerErr := schedulerService.Start(cmd.Context()); startSchedulerErr != nil {
		deps.Logger.Error("Failed to start scheduler service", "error", startSchedulerErr)
		return fmt.Errorf("failed to start scheduler service: %w", startSchedulerErr)
	}

	// Wait for interrupt signal or scheduler completion
	deps.Logger.Info("Waiting for interrupt signal")
	select {
	case <-done:
		// Scheduler completed (unlikely for continuous scheduler, but handle it)
		deps.Logger.Info("Scheduler completed")
	case <-cmd.Context().Done():
		// Interrupt signal received - graceful shutdown
		deps.Logger.Info("Shutdown signal received")
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.DefaultShutdownTimeout)
	defer cancel()

	// Stop the scheduler service
	if stopSchedulerErr := schedulerService.Stop(shutdownCtx); stopSchedulerErr != nil {
		deps.Logger.Error("Failed to stop scheduler service", "error", stopSchedulerErr)
		return fmt.Errorf("failed to stop scheduler service: %w", stopSchedulerErr)
	}

	deps.Logger.Info("Scheduler stopped successfully")
	return nil
}

// createCrawlerInstance creates a crawler instance with the given services.
// This is a helper function to consolidate crawler creation logic.
func createCrawlerInstance(
	log logger.Interface,
	cfg config.Interface,
	sourceManager sources.Interface,
	storageResult *cmdcommon.StorageResult,
	articleService articlespkg.Interface,
	pageService pagepkg.Interface,
) (crawler.Interface, error) {
	// Create event bus
	bus := events.NewEventBus(log)

	// Get crawler config
	crawlerCfg := cfg.GetCrawlerConfig()
	if crawlerCfg == nil {
		return nil, errors.New("crawler configuration is required")
	}

	// Create crawler using NewCrawlerWithParams
	crawlerResult, err := crawler.NewCrawlerWithParams(crawler.CrawlerParams{
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
