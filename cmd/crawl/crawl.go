// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	articlespkg "github.com/jonesrussell/gocrawl/internal/content/articles"
	pagepkg "github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	loggerpkg "github.com/jonesrussell/gocrawl/internal/logger"
	sourcespkg "github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Crawler handles the crawl operation
type Crawler struct {
	config        config.Interface
	logger        loggerpkg.Interface
	jobService    common.JobService
	sourceManager sourcespkg.Interface
	crawler       crawler.Interface
	done          chan struct{} // Channel to signal crawler completion
}

// NewCrawler creates a new crawler instance
func NewCrawler(
	config config.Interface,
	logger loggerpkg.Interface,
	jobService common.JobService,
	sourceManager sourcespkg.Interface,
	crawler crawler.Interface,
	done chan struct{},
) *Crawler {
	return &Crawler{
		config:        config,
		logger:        logger,
		jobService:    jobService,
		sourceManager: sourceManager,
		crawler:       crawler,
		done:          done,
	}
}

// Start begins the crawl operation
func (c *Crawler) Start(ctx context.Context) error {
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cmdcommon.DefaultShutdownTimeout)
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
			// Get dependencies from context using helper
			log, cfg, err := cmdcommon.GetDependencies(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get dependencies: %w", err)
			}

			// Construct dependencies directly using the same functions FX would use
			// This avoids creating a new FX app per command execution
			crawlerInstance, err := constructCrawlerDependencies(log, cfg, args[0], maxDepth)
			if err != nil {
				return fmt.Errorf("failed to construct crawler dependencies: %w", err)
			}

			return crawlerInstance.Start(cmd.Context())
		},
	}

	// Add --max-depth flag
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Override the max_depth setting from source configuration (0 means use source default)")

	return cmd
}

// SetupCollector creates and configures a new collector instance.
func SetupCollector(
	ctx context.Context,
	logger loggerpkg.Interface,
	indexManager types.IndexManager,
	sources sourcespkg.Interface,
	eventBus *events.EventBus,
	articleService articlespkg.Interface,
	pageService pagepkg.Interface,
	cfg *crawlerconfig.Config,
) (crawler.Interface, error) {
	// Create crawler instance using ProvideCrawler
	storage, ok := indexManager.(types.Interface)
	if !ok {
		return nil, errors.New("index manager does not implement types.Interface")
	}

	result, err := crawler.ProvideCrawler(struct {
		fx.In
		Logger         loggerpkg.Interface
		Bus            *events.EventBus
		IndexManager   types.IndexManager
		Sources        sourcespkg.Interface
		Config         *crawlerconfig.Config
		ArticleService articlespkg.Interface
		PageService    pagepkg.Interface
		Storage        types.Interface
	}{
		Logger:         logger,
		Bus:            eventBus,
		IndexManager:   indexManager,
		Sources:        sources,
		Config:         cfg,
		ArticleService: articleService,
		PageService:    pageService,
		Storage:        storage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create crawler: %w", err)
	}

	return result.Crawler, nil
}

// constructCrawlerDependencies constructs all dependencies needed for the crawl command
// without using FX, by directly calling the same constructors that FX modules use
// maxDepthOverride: if > 0, overrides the source's max_depth setting
func constructCrawlerDependencies(log loggerpkg.Interface, cfg config.Interface, sourceName string, maxDepthOverride int) (*Crawler, error) {
	// Load sources
	sourceManager, err := sourcespkg.LoadSources(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	// Create storage client and storage
	storageClient, err := createStorageClientForCrawl(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: storageClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create event bus
	bus := events.NewEventBus(log)

	// Get crawler config
	crawlerCfg := cfg.GetCrawlerConfig()
	if crawlerCfg == nil {
		return nil, fmt.Errorf("crawler configuration is required")
	}

	// Construct article and page services directly
	// Get index names from config or use defaults
	articleIndexName := "articles"
	pageIndexName := "pages"
	if source := sourceManager.FindByName(sourceName); source != nil {
		if source.ArticleIndex != "" {
			articleIndexName = source.ArticleIndex
		}
		if source.Index != "" {
			pageIndexName = source.Index
		}
	}

	articleService := articlespkg.NewContentService(log, storageResult.Storage, articleIndexName)
	pageService := pagepkg.NewContentService(log, storageResult.Storage, pageIndexName)

	// Use the crawler's ProvideCrawler to construct the crawler
	crawlerResult, err := crawler.ProvideCrawler(struct {
		fx.In
		Logger         loggerpkg.Interface
		Bus            *events.EventBus
		IndexManager   types.IndexManager
		Sources        sourcespkg.Interface
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

	// Override max depth if flag is provided
	if maxDepthOverride > 0 {
		log.Info("Overriding source max_depth with flag value", "max_depth", maxDepthOverride)
		crawlerResult.Crawler.SetMaxDepth(maxDepthOverride)
	}

	// Create done channel for job service
	done := make(chan struct{})

	// Create processor factory (simplified)
	processorFactory := crawler.NewProcessorFactory(log, storageResult.Storage, "content")

	// Create job service
	jobService := NewJobService(JobServiceParams{
		Logger:           log,
		Sources:          sourceManager,
		Crawler:          crawlerResult.Crawler,
		Done:             done,
		Storage:          storageResult.Storage,
		ProcessorFactory: processorFactory,
		SourceName:       sourceName,
	})

	// Create crawler command instance
	return NewCrawler(cfg, log, jobService, sourceManager, crawlerResult.Crawler, done), nil
}

// createStorageClientForCrawl creates an Elasticsearch client for the crawl command
func createStorageClientForCrawl(cfg config.Interface, log loggerpkg.Interface) (*es.Client, error) {
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: cfg,
		Logger: log,
	})
	if err != nil {
		return nil, err
	}
	return clientResult.Client, nil
}
