// Package indices implements the indices command.
package indices

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the indices command dependencies.
var Module = fx.Module("indices",
	config.Module,
	storage.Module,
	sources.Module,
	logger.Module,
	fx.Provide(
		func() logger.Params {
			return logger.Params{
				Config: &logger.Config{
					Level:       logger.InfoLevel,
					Development: true,
					Encoding:    "console",
				},
			}
		},
		NewCreator,
		NewLister,
		NewTableRenderer,
		NewDeleter,
	),
)

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
