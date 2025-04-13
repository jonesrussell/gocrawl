// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Cmd represents the scheduler command.
var Cmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Start the scheduler",
	Long: `Start the scheduler to manage and execute scheduled crawling tasks.
The scheduler will run continuously until interrupted with Ctrl+C.`,
	RunE: runScheduler,
}

// runScheduler executes the scheduler command
func runScheduler(cmd *cobra.Command, _ []string) error {
	// Get logger from context
	loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New("logger not found in context or invalid type")
	}

	// Get config from context
	configValue := cmd.Context().Value(cmdcommon.ConfigKey)
	cfg, ok := configValue.(config.Interface)
	if !ok {
		return errors.New("config not found in context or invalid type")
	}

	// Create source manager
	sourceManager, err := sources.LoadSources(cfg)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	var schedulerService common.JobService

	// Create Fx app with the module
	fxApp := fx.New(
		Module,
		fx.Provide(
			func() logger.Interface { return log },
			func() sources.Interface { return sourceManager },
		),
		fx.WithLogger(func() fxevent.Logger {
			return logger.NewFxLogger(log)
		}),
		fx.Invoke(func(js common.JobService) {
			schedulerService = js
		}),
	)

	// Start the application
	log.Info("Starting application")
	startErr := fxApp.Start(cmd.Context())
	if startErr != nil {
		log.Error("Failed to start application", "error", startErr)
		return fmt.Errorf("failed to start application: %w", startErr)
	}

	// Start the scheduler service
	if startSchedulerErr := schedulerService.Start(cmd.Context()); startSchedulerErr != nil {
		log.Error("Failed to start scheduler service", "error", startSchedulerErr)
		return fmt.Errorf("failed to start scheduler service: %w", startSchedulerErr)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

	// Stop the scheduler service
	if stopSchedulerErr := schedulerService.Stop(cmd.Context()); stopSchedulerErr != nil {
		log.Error("Failed to stop scheduler service", "error", stopSchedulerErr)
		return fmt.Errorf("failed to stop scheduler service: %w", stopSchedulerErr)
	}

	// Stop the application
	log.Info("Stopping application")
	stopErr := fxApp.Stop(cmd.Context())
	if stopErr != nil {
		log.Error("Failed to stop application", "error", stopErr)
		return fmt.Errorf("failed to stop application: %w", stopErr)
	}

	return nil
}

// Command returns the scheduler command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
