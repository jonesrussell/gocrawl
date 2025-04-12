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
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
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

	log.Debug("Crawl command starting",
		zap.String("source", sourceName),
		zap.Bool("debug", cmd.Flag("debug").Value.String() == "true"))
	log.Info("Setting up crawl",
		zap.String("source", sourceName),
		zap.Bool("debug", cmd.Flag("debug").Value.String() == "true"))

	// Set debug mode from configuration
	config := common.GetConfigFromContext(cmdCtx)
	if config == nil {
		return errors.New("configuration not found in context")
	}
	debug := config.GetBool("app.debug") || config.GetString("logger.level") == "debug"
	if debug {
		log.Debug("Debug mode enabled",
			zap.Bool("app.debug", config.GetBool("app.debug")),
			zap.String("logger.level", config.GetString("logger.level")),
			zap.Bool("config_development", config.GetBool("logger.development")))
	}

	// Log important configuration values once
	log.Info("Configuration values",
		"app.debug", config.GetBool("app.debug"),
		"logger.level", config.GetString("logger.level"),
		"logger.format", config.GetString("logger.format"),
		"logger.output", config.GetString("logger.output"),
		"crawler.max_depth", config.GetInt("crawler.max_depth"),
		"crawler.rate_limit", config.GetString("crawler.rate_limit"),
		"crawler.user_agent", config.GetString("crawler.user_agent"),
		"elasticsearch.url", config.GetString("elasticsearch.url"),
		"elasticsearch.index_prefix", config.GetString("elasticsearch.index_prefix"),
	)

	// Initialize the Fx application.
	log.Debug("Initializing Fx application with modules",
		zap.String("source", sourceName),
		zap.Bool("context_available", cmdCtx != nil),
		zap.Bool("logger_available", log != nil))

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
		fx.WithLogger(func(log logger.Interface) fxevent.Logger {
			// Create a new logger with debug level for Fx events
			fxLog := log.With(
				zap.String("component", "fx"),
				zap.Bool("debug", debug),
				zap.String("source", sourceName),
			)
			// Create Fx logger with debug level
			fxLogger := logger.NewFxLogger(fxLog)
			// Set debug level for Fx events
			if debug {
				fxLog.Debug("Fx logger configured with debug level")
			}
			return fxLogger
		}),
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
