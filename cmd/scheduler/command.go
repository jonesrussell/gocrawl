// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
)

// NewSchedulerSubCommands returns the scheduler subcommands.
func NewSchedulerSubCommands(log logger.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage crawl scheduler",
		Long: `The scheduler command provides functionality for managing crawl schedules.
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
func newScheduleCmd(log logger.Interface) *cobra.Command {
	return &cobra.Command{
		Use:   "schedule",
		Short: "Schedule a new crawl task",
		Long: `Schedule a new crawl task with the specified parameters.
The task will be executed according to the provided schedule.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Scheduling new task")
			return nil
		},
	}
}

// newListCmd creates the list command.
func newListCmd(log logger.Interface) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all scheduled crawls",
		Long: `List all scheduled and completed crawl tasks.
This command shows the status and details of each scheduled task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Listing scheduled tasks")
			return nil
		},
	}
}

// newDeleteCmd creates the delete command.
func newDeleteCmd(log logger.Interface) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a scheduled crawl",
		Long: `Delete a specific scheduled crawl by its ID.
This command will remove the task from the scheduler.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Deleting scheduled task")
			return nil
		},
	}
}
