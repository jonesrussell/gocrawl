// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"go.uber.org/fx"
)

// Module provides the search command dependencies
var Module = fx.Module("search",
	api.Module,
)
