// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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

// runCrawl is the main logic of the crawl command.
func runCrawl(cmd *cobra.Command, args []string) error {
	sourceName := strings.Trim(args[0], "\"")
	cmdCtx := cmd.Context()
	loggerValue := cmd.Root().Context().Value(common.LoggerKey)
	if loggerValue == nil {
		return errors.New("logger not found in context")
	}
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New("invalid logger type in context")
	}

	log.Info("Setting up crawl", "source", sourceName)

	// Create a new context for the crawler and Fx application.
	crawlCtx, cancel := context.WithCancel(cmdCtx)
	defer cancel()

	log.Debug("Created crawl context")

	// Set up signal handling.
	handler := signal.NewSignalHandler(log)
	cleanup := handler.Setup(crawlCtx)
	defer cleanup()

	log.Debug("Set up signal handling")

	// Initialize the Fx application.
	log.Debug("Initializing Fx application")
	fxApp := fx.New(
		Module,
		fx.Provide(
			fx.Annotate(
				func() context.Context { return crawlCtx },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return sourceName },
				fx.ResultTags(`name:"sourceName"`),
			),
		),
		fx.Invoke(func(config config.Interface) {
			// Set debug mode from command line flag
			config.GetAppConfig().Debug = cmd.Root().Flags().Lookup("debug").Value.String() == "true"
		}),
		fx.Invoke(func(crawlerSvc crawler.Interface) {
			// Start the crawler
			log.Debug("Starting crawler")
			if err := crawlerSvc.Start(crawlCtx, sourceName); err != nil {
				log.Error("Failed to start crawler", "error", err)
				handler.RequestShutdown()
				return
			}

			// Wait for crawler to complete or context cancellation
			go func() {
				// Wait for either the crawler to finish or context cancellation
				select {
				case <-crawlerSvc.Done():
					log.Info("Crawler finished processing")
				case <-crawlCtx.Done():
					log.Info("Crawler context cancelled")
					// Create a timeout context for stopping the crawler
					stopCtx, stopCancel := context.WithTimeout(context.Background(), crawler.DefaultStopTimeout)
					defer stopCancel()

					// Stop the crawler gracefully
					if err := crawlerSvc.Stop(stopCtx); err != nil {
						log.Error("Failed to stop crawler gracefully", "error", err)
					}
				}
				handler.RequestShutdown()
			}()
		}),
	)

	// Handle coordinated shutdown.
	handler.SetFXApp(fxApp)

	// Start the Fx application.
	log.Debug("Starting Fx application")
	if err := fxApp.Start(crawlCtx); err != nil {
		log.Error("Failed to start application", "error", err)
		return fmt.Errorf("failed to start application: %w", err)
	}
	log.Info("Fx application started successfully")

	// Wait for context cancellation
	log.Debug("Waiting for crawl completion or cancellation")
	<-crawlCtx.Done()
	log.Info("Crawl context finished", "error", crawlCtx.Err())
	handler.RequestShutdown()
	return handler.Wait()
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
