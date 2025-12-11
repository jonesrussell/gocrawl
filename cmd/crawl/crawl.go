// Package crawl implements the crawl command for fetching and processing web content.
package crawl

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
	"github.com/jonesrussell/gocrawl/internal/job"
	loggerpkg "github.com/jonesrussell/gocrawl/internal/logger"
	sourcespkg "github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/cobra"
)

// Crawler handles the crawl operation
type Crawler struct {
	config        config.Interface
	logger        loggerpkg.Interface
	jobService    job.Service
	sourceManager sourcespkg.Interface
	crawler       crawler.Interface
	done          chan struct{} // Channel to signal crawler completion
}

// NewCrawler creates a new crawler instance
func NewCrawler(
	cfg config.Interface,
	logger loggerpkg.Interface,
	jobService job.Service,
	sourceManager sourcespkg.Interface,
	crawlerInstance crawler.Interface,
	done chan struct{},
) *Crawler {
	return &Crawler{
		config:        cfg,
		logger:        logger,
		jobService:    jobService,
		sourceManager: sourceManager,
		crawler:       crawlerInstance,
		done:          done,
	}
}

// Start begins the crawl operation
func (c *Crawler) Start(ctx context.Context) error {
	// Note: The done channel is closed by the JobService, not here.
	// We don't need to clean it up here to avoid double-close panics.

	// Check if sources exist
	if _, err := sourcespkg.LoadSources(c.config, c.logger); err != nil {
		if errors.Is(err, loader.ErrNoSources) {
			c.logger.Info("No sources found in configuration. Please add sources to your config file.")
			c.logger.Info("You can use the 'sources list' command to view configured sources.")
			return nil
		}
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Start the job service
	if err := c.jobService.Start(ctx); err != nil {
		c.logger.Error("Failed to start job service", "error", err)
		return fmt.Errorf("failed to start job service: %w", err)
	}

	// Wait for either crawler completion or interrupt signal
	c.logger.Info("Crawling started, waiting for completion or interrupt...")

	select {
	case <-c.done:
		// Crawler completed successfully
		c.logger.Info("Crawler completed successfully")
		return nil
	case <-ctx.Done():
		// Interrupt signal received - graceful shutdown
		c.logger.Info("Shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.DefaultShutdownTimeout)
		defer cancel()

		// Stop the job service
		if err := c.jobService.Stop(shutdownCtx); err != nil {
			c.logger.Error("Failed to stop job service", "error", err)
			return fmt.Errorf("failed to stop job service: %w", err)
		}
		return ctx.Err()
	}
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	var maxDepth int

	cmd := &cobra.Command{
		Use:   "crawl [source]",
		Short: "Crawl a website for content",
		Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.

The --max-depth flag can be used to override the max_depth setting from the source configuration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get dependencies
			deps, err := cmdcommon.NewCommandDeps()
			if err != nil {
				return fmt.Errorf("failed to initialize dependencies: %w", err)
			}

			// Construct dependencies
			crawlerInstance, err := constructCrawlerDependencies(deps.Logger, deps.Config, args[0], maxDepth)
			if err != nil {
				return fmt.Errorf("failed to construct crawler dependencies: %w", err)
			}

			return crawlerInstance.Start(cmd.Context())
		},
	}

	// Add --max-depth flag
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0,
		"Override the max_depth setting from source configuration (0 means use source default)")

	return cmd
}

// getIndexNamesForSource returns the index names for a given source
func getIndexNamesForSource(sourceManager sourcespkg.Interface, sourceName string) (articleIndex, pageIndex string) {
	articleIndex = constants.DefaultArticleIndex
	pageIndex = constants.DefaultPageIndex

	if source := sourceManager.FindByName(sourceName); source != nil {
		if source.ArticleIndex != "" {
			articleIndex = source.ArticleIndex
		}
		if source.Index != "" {
			pageIndex = source.Index
		}
	}

	return articleIndex, pageIndex
}

// createCrawlerInstance creates a crawler instance with the given services.
// This is a helper function to consolidate crawler creation logic.
func createCrawlerInstance(
	log loggerpkg.Interface,
	cfg config.Interface,
	sourceManager sourcespkg.Interface,
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

// constructCrawlerDependencies constructs all dependencies needed for the crawl command.
// maxDepthOverride: if > 0, overrides the source's max_depth setting.
func constructCrawlerDependencies(
	log loggerpkg.Interface,
	cfg config.Interface,
	sourceName string,
	maxDepthOverride int,
) (*Crawler, error) {
	// Load sources
	sourceManager, err := sourcespkg.LoadSources(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	// Create storage
	storageResult, err := cmdcommon.CreateStorage(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Get index names for this source
	articleIndex, pageIndex := getIndexNamesForSource(sourceManager, sourceName)

	// Create article and page services
	articleService := articlespkg.NewContentServiceWithSources(
		log, storageResult.Storage, articleIndex, sourceManager)
	pageService := pagepkg.NewContentServiceWithSources(
		log, storageResult.Storage, pageIndex, sourceManager)

	// Create crawler
	crawlerInstance, err := createCrawlerInstance(
		log, cfg, sourceManager, storageResult, articleService, pageService)
	if err != nil {
		return nil, fmt.Errorf("failed to create crawler: %w", err)
	}

	// Override max depth if specified
	if maxDepthOverride > 0 {
		log.Info("Overriding source max_depth with flag value", "max_depth", maxDepthOverride)
		crawlerInstance.SetMaxDepth(maxDepthOverride)
	}

	// Create supporting services
	done := make(chan struct{})
	processorFactory := crawler.NewProcessorFactory(log, storageResult.Storage, constants.DefaultContentIndex)

	jobService := NewJobService(JobServiceParams{
		Logger:           log,
		Sources:          sourceManager,
		Crawler:          crawlerInstance,
		Done:             done,
		Storage:          storageResult.Storage,
		ProcessorFactory: processorFactory,
		SourceName:       sourceName,
	})

	return NewCrawler(cfg, log, jobService, sourceManager, crawlerInstance, done), nil
}
