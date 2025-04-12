// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// setupCrawlerConfig creates and validates the crawler configuration
func setupCrawlerConfig() *crawler.Config {
	return &crawler.Config{
		MaxDepth:    viper.GetInt("crawler.max_depth"),
		RateLimit:   viper.GetDuration("crawler.rate_limit"),
		Parallelism: viper.GetInt("crawler.parallelism"),
		UserAgent:   viper.GetString("crawler.user_agent"),
	}
}

// setupAppConfig creates the application configuration
func setupAppConfig() *app.Config {
	return &app.Config{
		Name:        viper.GetString("app.name"),
		Version:     viper.GetString("app.version"),
		Environment: viper.GetString("app.environment"),
		Debug:       viper.GetBool("app.debug"),
	}
}

// runCrawl executes the crawl command
func runCrawl(cmd *cobra.Command, args []string) error {
	// Get logger from context
	loggerValue := cmd.Context().Value(common.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New("logger not found in context or invalid type")
	}

	// Setup configurations
	crawlerConfig := setupCrawlerConfig()
	appConfig := setupAppConfig()

	// Create Fx app
	fxApp := fx.New(
		fx.Provide(
			func() *crawler.Config { return crawlerConfig },
			func() *app.Config { return appConfig },
			func() logger.Interface { return log },
		),
		crawler.Module,
	)

	// Start the application
	startErr := fxApp.Start(cmd.Context())
	if startErr != nil {
		return fmt.Errorf("failed to start application: %w", startErr)
	}

	// Wait for interrupt signal
	<-cmd.Context().Done()

	// Stop the application
	stopErr := fxApp.Stop(cmd.Context())
	if stopErr != nil {
		return fmt.Errorf("failed to stop application: %w", stopErr)
	}

	return nil
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
