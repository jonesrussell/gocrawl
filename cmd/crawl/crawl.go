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

	// Initialize the Fx application.
	log.Debug("Initializing Fx application")
	var handler signal.Interface
	fxApp := fx.New(
		Module,
		fx.Provide(
			fx.Annotate(
				func() context.Context { return cmdCtx },
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
		// Suppress Fx's default logging and use our logger format
		fx.WithLogger(func(log logger.Interface) fxevent.Logger {
			return log.NewFxLogger()
		}),
		fx.Invoke(func(lc fx.Lifecycle, crawlerSvc crawler.Interface, h signal.Interface) {
			handler = h
			// Register lifecycle hooks
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Debug("Starting crawler")
					if err := crawlerSvc.Start(ctx, sourceName); err != nil {
						log.Error("Failed to start crawler", "error", err)
						return err
					}

					// Monitor crawler completion in background
					go func() {
						select {
						case <-crawlerSvc.Done():
							log.Info("Crawler finished processing")
							handler.RequestShutdown()
						case <-ctx.Done():
							log.Info("Crawler context cancelled")
							// Create a timeout context for stopping the crawler
							stopCtx, stopCancel := context.WithTimeout(context.Background(), crawler.DefaultStopTimeout)
							defer stopCancel()

							// Stop the crawler gracefully
							if err := crawlerSvc.Stop(stopCtx); err != nil {
								log.Error("Failed to stop crawler gracefully", "error", err)
							}
							handler.RequestShutdown()
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Info("Stopping crawler")
					// Create a timeout context for stopping the crawler
					stopCtx, stopCancel := context.WithTimeout(context.Background(), crawler.DefaultStopTimeout)
					defer stopCancel()

					// Stop the crawler gracefully
					if err := crawlerSvc.Stop(stopCtx); err != nil {
						log.Error("Failed to stop crawler gracefully", "error", err)
						return err
					}
					return nil
				},
			})
		}),
	)

	// Start the Fx application.
	log.Debug("Starting Fx application")
	if err := fxApp.Start(cmdCtx); err != nil {
		log.Error("Failed to start application", "error", err)
		return fmt.Errorf("failed to start application: %w", err)
	}
	log.Info("Fx application started successfully")

	// Wait for shutdown signal
	if err := handler.Wait(); err != nil {
		log.Error("Error during shutdown", "error", err)
		return fmt.Errorf("shutdown error: %w", err)
	}

	// Stop the Fx application
	log.Info("Stopping Fx application")
	if err := fxApp.Stop(cmdCtx); err != nil {
		log.Error("Failed to stop application", "error", err)
		return fmt.Errorf("failed to stop application: %w", err)
	}

	return nil
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
