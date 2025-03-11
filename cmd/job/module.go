// Package job implements the job scheduler command.
package job

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the job command's dependencies.
var Module = fx.Module("job",
	fx.Provide(
		Command,
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"crawlDone"`),
		),
		fx.Annotate(
			func(cmd *cobra.Command) context.Context {
				return cmd.Context()
			},
			fx.ResultTags(`name:"jobContext"`),
		),
		func() chan *models.Article {
			return make(chan *models.Article, DefaultChannelBufferSize)
		},
	),
)
