// Package job provides the job scheduler implementation.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the job scheduler dependencies.
var Module = fx.Module("job",
	fx.Provide(
		// Provide the job scheduler
		func(
			sources sources.Interface,
			crawler crawler.Interface,
			logger common.Logger,
			done chan struct{},
		) Interface {
			return New(sources, crawler, logger, done)
		},
	),
)
