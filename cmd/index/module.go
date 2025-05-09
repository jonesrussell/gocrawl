// Package index provides commands for managing Elasticsearch index.
package index

import (
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Command is the index command
var Command = &cobra.Command{
	Use:   "index",
	Short: "Manage Elasticsearch indices",
	Long:  `Manage Elasticsearch indices for storing crawled content`,
}

// Module provides the index module for dependency injection.
var Module = fx.Module("index",
	common.Module,
	fx.Provide(
		// Provide the table renderer
		NewTableRenderer,

		// Provide the lister
		NewLister,

		// Provide the command
		func(
			cfg config.Interface,
			log logger.Interface,
			storage types.Interface,
			lister *Lister,
		) *cobra.Command {
			// Add subcommands
			Command.AddCommand(
				&cobra.Command{
					Use:   "list",
					Short: "List all indices",
					RunE: func(cmd *cobra.Command, args []string) error {
						return lister.Start(cmd.Context())
					},
				},
				&cobra.Command{
					Use:   "create",
					Short: "Create an index",
					RunE: func(cmd *cobra.Command, args []string) error {
						// TODO: Implement create functionality
						return nil
					},
				},
				&cobra.Command{
					Use:   "delete",
					Short: "Delete an index",
					RunE: func(cmd *cobra.Command, args []string) error {
						// TODO: Implement delete functionality
						return nil
					},
				},
			)

			return Command
		},
	),
)
