// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/content/articles"
	"github.com/jonesrussell/gocrawl/internal/content/page"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	loggerpkg "github.com/jonesrussell/gocrawl/internal/logger"
	sourcespkg "github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Crawler handles the crawl operation
type Crawler struct {
	config        config.Interface
	logger        loggerpkg.Interface
	jobService    common.JobService
	sourceManager sourcespkg.Interface
	crawler       crawler.Interface
}

// NewCrawler creates a new crawler instance
func NewCrawler(
	config config.Interface,
	logger loggerpkg.Interface,
	jobService common.JobService,
	sourceManager sourcespkg.Interface,
	crawler crawler.Interface,
) *Crawler {
	return &Crawler{
		config:        config,
		logger:        logger,
		jobService:    jobService,
		sourceManager: sourceManager,
		crawler:       crawler,
	}
}

// Start begins the crawl operation
func (c *Crawler) Start(ctx context.Context) error {
	// Check if sources exist
	if _, err := sourcespkg.LoadSources(c.config); err != nil {
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

	// Wait for interrupt signal
	c.logger.Info("Waiting for interrupt signal")
	<-ctx.Done()

	// Stop the job service
	if err := c.jobService.Stop(ctx); err != nil {
		c.logger.Error("Failed to stop job service", "error", err)
		return fmt.Errorf("failed to stop job service: %w", err)
	}

	return nil
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawl [source]",
		Short: "Crawl a website for content",
		Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(loggerpkg.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config from context
			configValue := cmd.Context().Value(cmdcommon.ConfigKey)
			cfg, ok := configValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context or invalid type")
			}

			// Create Fx application
			app := fx.New(
				// Include required modules
				Module,

				// Provide existing config
				fx.Provide(func() config.Interface { return cfg }),

				// Provide existing logger
				fx.Provide(func() loggerpkg.Interface { return log }),

				// Provide source name
				fx.Provide(fx.Annotate(
					func() string { return args[0] },
					fx.ResultTags(`name:"sourceName"`),
				)),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return loggerpkg.NewFxLogger(log)
				}),

				// Invoke crawler
				fx.Invoke(func(c *Crawler) error {
					return c.Start(cmd.Context())
				}),
			)

			// Start application
			if err := app.Start(cmd.Context()); err != nil {
				return fmt.Errorf("failed to start application: %w", err)
			}

			// Stop application
			if err := app.Stop(cmd.Context()); err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			return nil
		},
	}

	return cmd
}

// SetupCollector creates and configures a new collector instance.
func SetupCollector(
	ctx context.Context,
	logger loggerpkg.Interface,
	indexManager types.IndexManager,
	sources sourcespkg.Interface,
	eventBus *events.EventBus,
	articleService articles.Interface,
	pageService page.Interface,
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
		ArticleService articles.Interface
		PageService    page.Interface
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
