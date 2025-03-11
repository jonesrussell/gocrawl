// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// runScheduler manages the execution of scheduled jobs.
func runScheduler(ctx context.Context, logger common.Logger, sources *sources.Sources, rootCmd string) {
	logger.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Do initial check
	checkAndRunJobs(logger, sources, rootCmd, time.Now())

	for {
		select {
		case <-ctx.Done():
			logger.Info("Job scheduler shutting down")
			return
		case t := <-ticker.C:
			checkAndRunJobs(logger, sources, rootCmd, t)
		}
	}
}

// checkAndRunJobs evaluates and executes scheduled jobs.
func checkAndRunJobs(logger common.Logger, sources *sources.Sources, rootCmd string, now time.Time) {
	if sources == nil {
		logger.Error("Sources configuration is nil")
		return
	}

	currentTime := now.Format("15:04")
	logger.Info("Checking jobs", "current_time", currentTime)

	for _, source := range sources.Sources {
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				logger.Info("Running scheduled crawl",
					"source", source.Name,
					"time", scheduledTime)

				if err := runCrawlCommand(rootCmd, source.Name); err != nil {
					logger.Error("Error running crawl command",
						"error", err,
						"source", source.Name)
				}
			}
		}
	}
}

// runCrawlCommand executes a crawl command for a specific source.
func runCrawlCommand(rootCmd, sourceName string) error {
	cmd := exec.Command(rootCmd, "crawl", sourceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// loadSources loads the sources configuration from sources.yml
func loadSources() (*sources.Sources, error) {
	data, err := os.ReadFile("sources.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read sources.yml: %w", err)
	}

	var s sources.Sources
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse sources.yml: %w", err)
	}

	return &s, nil
}

// Cmd represents the job scheduler command.
var Cmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl jobs",
	Long:  `Schedule and run crawl jobs based on the times specified in sources.yml`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Initialize logger
		l, err := logger.NewDevelopmentLogger("info")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}

		// Load sources
		sources, err := loadSources()
		if err != nil {
			l.Error("Failed to load sources", "error", err)
			os.Exit(1)
		}

		// Print loaded schedules
		l.Info("Loaded schedules:")
		for _, source := range sources.Sources {
			if len(source.Time) > 0 {
				l.Info("Source schedule", "name", source.Name, "times", source.Time)
			}
		}

		// Create cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Start scheduler in background
		go runScheduler(ctx, l, sources, cmd.Root().Name())

		// Wait for interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		l.Info("Job scheduler running. Press Ctrl+C to stop...")
		<-sigChan
		l.Info("Shutting down...")
	},
}

// Command returns the job command.
func Command() *cobra.Command {
	return Cmd
}
