// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/common"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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

	var jobService common.JobService

	// Create Fx app with the module
	fxApp := fx.New(
		Module,
		fx.Provide(
			func() logger.Interface { return log },
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
	if err := jobService.Start(cmd.Context()); err != nil {
		log.Error("Failed to start job service", "error", err)
		return fmt.Errorf("failed to start job service: %w", err)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

	// Stop the job service
	if err := jobService.Stop(cmd.Context()); err != nil {
		log.Error("Failed to stop job service", "error", err)
		return fmt.Errorf("failed to stop job service: %w", err)
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
	indexManager interfaces.IndexManager,
	sources sources.Interface,
	eventBus *events.EventBus,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	cfg *crawlerconfig.Config,
) (crawler.Interface, error) {
	// Create collector with rate limiting
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// Create crawler instance
	crawler := crawler.NewCrawler(
		logger,
		eventBus,
		indexManager,
		sources,
		articleProcessor,
		contentProcessor,
		cfg,
	)

	// Set up event handling
	eventHandler := events.NewDefaultHandler(logger)
	eventBus.Subscribe(eventHandler)

	return crawler, nil
}
