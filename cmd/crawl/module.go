// Package crawl implements the crawl command.
package crawl

import (
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/models"
	"go.uber.org/fx"
)

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	fx.Provide(
		// Core dependencies
		func() chan struct{} {
			return make(chan struct{})
		},
		func() chan *models.Article {
			return make(chan *models.Article, app.DefaultChannelBufferSize)
		},
	),
)
