package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
	}
	appInstance *fx.App
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	var lgr *logger.CustomLogger

	appInstance = fx.New(
		// Core modules
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,

		// Application module
		app.Module,

		fx.Populate(&lgr),

		// Provide base context
		fx.Provide(func() context.Context {
			return context.Background()
		}),

		// Provide crawler constructor
		fx.Provide(
			func(cfg *config.Config, log logger.Interface, store storage.Interface) (*crawler.Crawler, error) {
				params := crawler.Params{
					BaseURL:   cfg.Crawler.BaseURL,
					MaxDepth:  cfg.Crawler.MaxDepth,
					RateLimit: cfg.Crawler.RateLimit,
					Debugger:  &logger.CollyDebugger{Logger: log},
					Logger:    log,
					Config:    cfg,
					Storage:   store,
				}
				result, err := crawler.NewCrawler(params)
				if err != nil {
					log.Error("Failed to create crawler", "error", err)
					return nil, err
				}
				log.Debug("Crawler created successfully", "baseURL", params.BaseURL)
				return result.Crawler, nil
			},
		),
	)

	// Create a context for the application and add the logger to it
	ctx := logger.WithContext(context.Background(), lgr.GetZapLogger()) // Use GetZapLogger() to get the zap.Logger

	if err := appInstance.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(NewCrawlCmd(lgr)) // Pass logger from appInstance
	rootCmd.AddCommand(NewSearchCmd(lgr))

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	if err := appInstance.Stop(ctx); err != nil {
		return fmt.Errorf("error stopping application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	if err := appInstance.Stop(ctx); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	return nil
}
