// Package crawl implements the crawl command.
package crawl

import (
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"go.uber.org/fx"
)

// Module provides the crawl command's dependencies.
var Module = fx.Module("crawl",
	crawler.Module,
)
