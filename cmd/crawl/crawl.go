// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// processorTimeout is the timeout for each processor
	processorTimeout = 30 * time.Second
	// crawlerTimeout is the timeout for waiting for the crawler to complete
	crawlerTimeout = 5 * time.Minute
	// shutdownTimeout is the timeout for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

// Crawler implements the crawler command
type Crawler struct {
	config      config.Interface
	logger      logger.Interface
	storage     storagetypes.Interface
	crawler     crawler.Interface
	processors  []common.Processor
	articleChan chan *models.Article
	sourceName  string
}

// NewCrawler creates a new crawler instance
func NewCrawler(
	config config.Interface,
	logger logger.Interface,
	storage storagetypes.Interface,
	crawler crawler.Interface,
	processors []common.Processor,
	articleChan chan *models.Article,
	sourceName string,
) *Crawler {
	return &Crawler{
		config:      config,
		logger:      logger,
		storage:     storage,
		crawler:     crawler,
		processors:  processors,
		articleChan: articleChan,
		sourceName:  sourceName,
	}
}

// Start starts the crawler
func (c *Crawler) Start(ctx context.Context) error {
	// Test storage connection
	if err := c.storage.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}

	// Start crawler
	c.logger.Info("Starting crawler...", "source", c.sourceName)
	if err := c.crawler.Start(ctx, c.sourceName); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Process articles from the channel
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping article processing")
				return
			case article, ok := <-c.articleChan:
				if !ok {
					c.logger.Info("Article channel closed")
					return
				}

				// Create a job for each article with timeout
				job := &common.Job{
					ID:        article.ID,
					URL:       article.Source,
					Status:    "pending",
					CreatedAt: article.CreatedAt,
					UpdatedAt: article.UpdatedAt,
				}

				// Process the article using each processor with timeout
				for _, processor := range c.processors {
					// Create a timeout context for each processor
					processorCtx, processorCancel := context.WithTimeout(ctx, processorTimeout)
					processor.ProcessJob(processorCtx, job)
					processorCancel()
				}
			}
		}
	}()

	return nil
}

// Stop gracefully stops the crawler
func (c *Crawler) Stop(ctx context.Context) error {
	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Stop crawler
	if err := c.crawler.Stop(shutdownCtx); err != nil {
		if !strings.Contains(err.Error(), "Max depth limit reached") {
			return fmt.Errorf("failed to stop crawler: %w", err)
		}
		c.logger.Info("Crawler stopped at max depth limit")
	}

	// Close storage connection
	if err := c.storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage connection: %w", err)
	}

	return nil
}

// Cmd represents the crawl command
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Trim quotes from source name
		sourceName := strings.Trim(args[0], "\"")

		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Set up signal handling with a no-op logger initially
		handler := signal.NewSignalHandler(logger.NewNoOp())
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Initialize the Fx application
		fxApp := fx.New(
			fx.Provide(
				func() context.Context { return ctx },
				func() string { return sourceName },
			),
			fx.Invoke(func(lc fx.Lifecycle, crawler *Crawler) {
				// Update the signal handler with the real logger
				handler.SetLogger(crawler.logger)

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						if err := crawler.Start(ctx); err != nil {
							return err
						}

						// Wait for crawler to complete with timeout
						go func() {
							// Create a timeout context for waiting
							waitCtx, waitCancel := context.WithTimeout(ctx, crawlerTimeout)
							defer waitCancel()

							select {
							case <-waitCtx.Done():
								crawler.logger.Info("Crawler reached timeout limit")
							default:
								crawler.crawler.Wait()
								crawler.logger.Info("Crawler finished processing")
							}
							// Signal completion to the signal handler
							handler.RequestShutdown()
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						return crawler.Stop(ctx)
					},
				})
			}),
		)

		// Set the fx app for coordinated shutdown
		handler.SetFXApp(fxApp)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("failed to start application: %w", err)
		}

		// Wait for completion signal
		handler.Wait()

		return nil
	},
}

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
