// Package indexing provides functionality for managing Elasticsearch indices.
package indexing

import (
	"go.uber.org/fx"
)

// Module provides indexing dependencies
var Module = fx.Module("indexing",
	fx.Provide(),
)
