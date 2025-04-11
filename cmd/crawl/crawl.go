// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
			// Include all required modules
			config.Module,
			storage.Module,
			logger.Module,
			crawler.Module,
			sources.Module,
			Module, // Include the crawl command module
			// Provide context and source name
			fx.Provide(
				func() context.Context { return ctx },
				func() string { return sourceName },
			),
			// Provide logger params
			fx.Provide(func() logger.Params {
				return logger.Params{
					Config: &logger.Config{
						Level:       logger.InfoLevel,
						Development: true,
						Encoding:    "console",
					},
				}
			}),
			// Provide article channel
			fx.Provide(func() chan *models.Article {
				return make(chan *models.Article, 100)
			}),
			// Provide processors
			fx.Provide(func() []common.Processor {
				return []common.Processor{
					// Add your processors here
				}
			}),
			// Provide the event bus
			fx.Provide(events.NewBus),
			// Provide the IndexManager
			fx.Provide(func(client *elasticsearch.Client, logger logger.Interface) interfaces.IndexManager {
				return storage.NewElasticsearchIndexManager(client, logger)
			}),
			fx.Invoke(func(lc fx.Lifecycle, crawler *CrawlerCommand) {
				// Update the signal handler with the real logger
				handler.SetLogger(crawler.logger)

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Start the crawler
						if err := crawler.Start(ctx, sourceName); err != nil {
							return fmt.Errorf("failed to start crawler: %w", err)
						}

						// Start a goroutine to wait for crawler completion
						go func() {
							// Create a timeout context for waiting
							waitCtx, waitCancel := context.WithTimeout(ctx, crawlerTimeout)
							defer waitCancel()

							// Wait for crawler to complete
							crawler.Wait()

							// Check if we timed out
							select {
							case <-waitCtx.Done():
								crawler.logger.Info("Crawler reached timeout limit")
							default:
								crawler.logger.Info("Crawler finished processing")
							}

							// Signal completion to the signal handler
							handler.RequestShutdown()
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						// Stop the crawler with timeout
						stopCtx, stopCancel := context.WithTimeout(ctx, shutdownTimeout)
						defer stopCancel()

						if err := crawler.Stop(stopCtx); err != nil {
							return fmt.Errorf("failed to stop crawler: %w", err)
						}
						return nil
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
