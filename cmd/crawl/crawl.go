// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
			Module,
			// storage.Module is already provided through app.Module
			fx.Provide(
				fx.Annotate(
					func() context.Context { return ctx },
					fx.ResultTags(`name:"crawlContext"`),
				),
				fx.Annotate(
					func() string { return sourceName },
					fx.ResultTags(`name:"sourceName"`),
				),
			),
			fx.Invoke(func(lc fx.Lifecycle, p CommandDeps) {
				// Update the signal handler with the real logger
				handler.SetLogger(p.Logger)

				// Create a channel to signal when the crawler is done
				crawlerDone := make(chan struct{})

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						// Test storage connection
						if testErr := p.Storage.TestConnection(p.Context); testErr != nil {
							return fmt.Errorf("failed to connect to storage: %w", testErr)
						}

						// Start crawler
						p.Logger.Info("Starting crawler...", "source", p.SourceName)
						if startErr := p.Crawler.Start(p.Context, p.SourceName); startErr != nil {
							return fmt.Errorf("failed to start crawler: %w", startErr)
						}

						// Process articles from the channel
						go func() {
							defer close(crawlerDone)
							for {
								select {
								case <-ctx.Done():
									p.Logger.Info("Context cancelled, stopping article processing")
									return
								case article, ok := <-p.ArticleChan:
									if !ok {
										p.Logger.Info("Article channel closed")
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
									for _, processor := range p.Processors {
										// Create a timeout context for each processor
										processorCtx, processorCancel := context.WithTimeout(ctx, processorTimeout)
										processor.ProcessJob(processorCtx, job)
										processorCancel()
									}
								}
							}
						}()

						// Wait for crawler to complete with timeout
						go func() {
							// Create a timeout context for waiting
							waitCtx, waitCancel := context.WithTimeout(ctx, crawlerTimeout)
							defer waitCancel()

							select {
							case <-waitCtx.Done():
								p.Logger.Info("Crawler reached timeout limit")
							default:
								p.Crawler.Wait()
								p.Logger.Info("Crawler finished processing")
							}
							// Signal completion to the signal handler
							handler.RequestShutdown()
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						// Create a timeout context for shutdown
						shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
						defer shutdownCancel()

						// Stop crawler
						if stopErr := p.Crawler.Stop(shutdownCtx); stopErr != nil {
							if !strings.Contains(stopErr.Error(), "Max depth limit reached") {
								return fmt.Errorf("failed to stop crawler: %w", stopErr)
							}
							p.Logger.Info("Crawler stopped at max depth limit")
						}

						// Wait for article processing to complete
						select {
						case <-crawlerDone:
							p.Logger.Info("Article processing completed")
						case <-shutdownCtx.Done():
							p.Logger.Info("Shutdown timeout reached, stopping article processing")
						}

						// Close storage connection
						if closeErr := p.Storage.Close(); closeErr != nil {
							return fmt.Errorf("failed to close storage connection: %w", closeErr)
						}

						return nil
					},
				})
			}),
		)

		// Set the fx app for coordinated shutdown
		handler.SetFXApp(fxApp)

		// Start the application
		if startErr := fxApp.Start(ctx); startErr != nil {
			return fmt.Errorf("failed to start application: %w", startErr)
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
