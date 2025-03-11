// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the crawl command dependencies
var Module = fx.Module("crawl",
	crawler.Module,
	sources.Module,
)
