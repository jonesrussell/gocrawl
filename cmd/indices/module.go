// Package indices provides commands for managing Elasticsearch indices.
package indices

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the indices module for dependency injection.
var Module = fx.Module("indices",
	fx.Provide(
		NewCreator,
		NewLister,
		NewTableRenderer,
		NewDeleter,
	),
)

// RegisterCommands registers the indices commands with the root command.
func RegisterCommands(rootCmd *cobra.Command) {
	indicesCmd := &cobra.Command{
		Use:   "indices",
		Short: "Manage Elasticsearch indices",
	}

	indicesCmd.AddCommand(
		NewListCommand(),
		NewCreateCommand(),
		NewDeleteCommand(),
	)

	rootCmd.AddCommand(indicesCmd)
}

// NewIndices creates a new indices command.
func NewIndices(p struct {
	fx.In
	Context context.Context `name:"indicesContext"`
	Config  config.Interface
	Logger  logger.Interface
	Storage types.Interface
}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indices",
		Short: "Manage Elasticsearch indices",
		Long:  `Manage Elasticsearch indices for the crawler.`,
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewCreateCommand())

	return cmd
}
