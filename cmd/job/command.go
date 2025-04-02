// Package job provides the job command implementation.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
)

// NewJobSubCommands returns the job subcommands.
func NewJobSubCommands(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage crawl jobs",
		Long: `The job command provides functionality for managing crawl jobs.
It allows you to schedule, list, and manage web crawling tasks.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newScheduleCmd(log),
		newListCmd(log),
		newDeleteCmd(log),
	)

	return cmd
}

// newScheduleCmd creates the schedule command.
func newScheduleCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "schedule",
		Short: "Schedule a new crawl job",
		Long: `Schedule a new crawl job with the specified parameters.
The job will be executed according to the provided schedule.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Scheduling new job")
			return nil
		},
	}
}

// newListCmd creates the list command.
func newListCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all crawl jobs",
		Long: `List all scheduled and completed crawl jobs.
This command shows the status and details of each job.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Listing jobs")
			return nil
		},
	}
}

// newDeleteCmd creates the delete command.
func newDeleteCmd(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a crawl job",
		Long: `Delete a specific crawl job by its ID.
This command will remove the job from the scheduler.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Deleting job")
			return nil
		},
	}
}
