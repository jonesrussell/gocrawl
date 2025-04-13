// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
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

// runCrawl executes the crawl command
func runCrawl(cmd *cobra.Command, args []string) error {
	// Get logger from context
	loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New("logger not found in context or invalid type")
	}

	// Create Fx app with the module
	fxApp := fx.New(
		Module,
		fx.Provide(
			func() logger.Interface { return log },
			func() string { return args[0] }, // Provide source name
		),
		fx.WithLogger(func() fxevent.Logger {
			return logger.NewFxLogger(log)
		}),
	)

	// Start the application
	log.Info("Starting application")
	startErr := fxApp.Start(cmd.Context())
	if startErr != nil {
		log.Error("Failed to start application", "error", startErr)
		return fmt.Errorf("failed to start application: %w", startErr)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

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
