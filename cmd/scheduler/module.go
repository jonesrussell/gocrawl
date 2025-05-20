// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	internalcommon "github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the scheduler command module for dependency injection
var Module = fx.Module("scheduler",
	// Include required modules
	cmdcommon.Module,
	crawler.Module,

	// Provide the scheduler service
	fx.Provide(
		// Provide the done channel
		func() chan struct{} {
			return make(chan struct{})
		},

		// Provide the active jobs counter
		func() *int32 {
			var jobs int32
			return &jobs
		},

		// Provide the scheduler service
		fx.Annotate(
			NewSchedulerService,
			fx.As(new(internalcommon.JobService)),
		),

		// Provide the command registrar
		fx.Annotated{
			Group: "commands",
			Target: func(
				deps cmdcommon.CommandDeps,
				service internalcommon.JobService,
			) cmdcommon.CommandRegistrar {
				return func(parent *cobra.Command) {
					cmd := &cobra.Command{
						Use:   "scheduler",
						Short: "Manage scheduled crawling tasks",
						Long:  `Manage scheduled crawling tasks for content sources`,
					}

					// Add subcommands
					cmd.AddCommand(
						&cobra.Command{
							Use:   "start",
							Short: "Start the scheduler",
							RunE: func(cmd *cobra.Command, args []string) error {
								return service.Start(cmd.Context())
							},
						},
						&cobra.Command{
							Use:   "stop",
							Short: "Stop the scheduler",
							RunE: func(cmd *cobra.Command, args []string) error {
								return service.Stop(cmd.Context())
							},
						},
					)

					parent.AddCommand(cmd)
				}
			},
		},
	),
)
