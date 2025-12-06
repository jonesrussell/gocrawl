// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"github.com/jonesrussell/gocrawl/cmd/common"
	"go.uber.org/fx"
)

// Module provides the search command dependencies.
// Note: Command registration is handled by Command() function in search.go, not through FX Group annotation.
var Module = fx.Module("search",
	common.Module,
)
