// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
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
	log := common.GetLoggerFromContext(cmdCtx)
	if log == nil {
		return errors.New("logger not found in context")
	}

	log.Info("Setting up crawl", "source", sourceName)

	// Set debug mode from configuration
	config := common.GetConfigFromContext(cmdCtx)
	if config == nil {
		return errors.New("configuration not found in context")
	}
	debug := config.GetBool("app.debug") || config.GetString("logger.level") == "debug"
	if debug {
		log.Debug("Debug mode enabled")
		log.Debug("Configuration loaded",
			"max_depth", config.GetInt("crawler.max_depth"),
			"rate_limit", config.GetString("crawler.rate_limit"),
			"user_agent", config.GetString("crawler.user_agent"),
		)
	}

	// Initialize the Fx application.
	log.Debug("Initializing Fx application with modules",
		"source", sourceName,
		"context_available", cmdCtx != nil,
		"logger_available", log != nil,
	)

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
			fx.Annotate(
				func() logger.Interface { return log },
				fx.As(new(logger.Interface)),
			),
		),
		// Configure Fx logger with debug level and more detailed logging
		// fx.WithLogger(func(log logger.Interface) fxevent.Logger {
		// 	// Create a new logger with debug level for Fx events
		// 	fxLog := log.With(
		// 		"component", "fx",
		// 		"debug", debug,
		// 		"source", sourceName,
		// 	)
		// 	// Log initialization start with structured fields
		// 	fxLog.Debug("Starting Fx application initialization",
		// 		"source", sourceName,
		// 		"debug_enabled", debug,
		// 		"context_available", cmdCtx != nil,
		// 	)
		// 	// Create Fx logger with debug level
		// 	fxLogger := logger.NewFxLogger(fxLog)
		// 	// Log initial Fx event to verify logging
		// 	fxLog.Debug("Fx logger initialized",
		// 		"source", sourceName,
		// 		"debug_enabled", debug,
		// 	)
		// 	return fxLogger
		// }),
		fx.Invoke(func(lc fx.Lifecycle, crawlerSvc crawler.Interface, h signal.Interface) {
			handler = h
			// Register lifecycle hooks
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Debug("Starting crawler lifecycle",
						"source", sourceName,
						"context_available", ctx != nil,
						"crawler_available", crawlerSvc != nil,
					)

					if err := crawlerSvc.Start(ctx, sourceName); err != nil {
						log.Error("Failed to start crawler", "error", err)
						return err
					}

					// Monitor crawler completion in background
					go func() {
						log.Debug("Starting crawler monitoring goroutine")
						select {
						case <-crawlerSvc.Done():
							log.Info("Crawler finished processing", "source", sourceName)
							handler.RequestShutdown()
						case <-ctx.Done():
							log.Info("Crawler context cancelled", "source", sourceName)
							// Create a timeout context for stopping the crawler
							stopCtx, stopCancel := context.WithTimeout(context.Background(), crawler.DefaultStopTimeout)
							defer stopCancel()

							// Stop the crawler gracefully
							if err := crawlerSvc.Stop(stopCtx); err != nil {
								log.Error("Failed to stop crawler gracefully",
									"error", err,
									"source", sourceName,
								)
							}
							handler.RequestShutdown()
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Debug("Stopping crawler lifecycle", "source", sourceName)
					// Create a timeout context for stopping the crawler
					stopCtx, stopCancel := context.WithTimeout(context.Background(), crawler.DefaultStopTimeout)
					defer stopCancel()

					// Stop the crawler gracefully
					if err := crawlerSvc.Stop(stopCtx); err != nil {
						log.Error("Failed to stop crawler gracefully",
							"error", err,
							"source", sourceName,
						)
						return err
					}
					return nil
				},
			})
		}),
	)

	// Start the Fx application.
	log.Debug("Starting Fx application", "source", sourceName)
	if err := fxApp.Start(cmdCtx); err != nil {
		log.Error("Failed to start application",
			"error", err,
			"source", sourceName,
		)
		return fmt.Errorf("failed to start application: %w", err)
	}
	log.Info("Fx application started successfully", "source", sourceName)

	// Wait for shutdown signal
	log.Debug("Waiting for shutdown signal", "source", sourceName)
	if err := handler.Wait(); err != nil {
		log.Error("Error during shutdown",
			"error", err,
			"source", sourceName,
		)
		return fmt.Errorf("shutdown error: %w", err)
	}

	// Stop the Fx application
	log.Debug("Stopping Fx application", "source", sourceName)
	if err := fxApp.Stop(cmdCtx); err != nil {
		log.Error("Failed to stop application",
			"error", err,
			"source", sourceName,
		)
		return fmt.Errorf("failed to stop application: %w", err)
	}

	log.Info("Crawl completed successfully", "source", sourceName)
	return nil
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
