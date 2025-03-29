// Package job implements the job scheduler command.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/models"
	"go.uber.org/fx"
)

// Module provides the job command's dependencies.
var Module = fx.Module("job",
	fx.Provide(
		// Core dependencies
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"crawlDone"`),
		),
		func() chan *models.Article {
			return make(chan *models.Article, app.DefaultChannelBufferSize)
		},
	),
)
