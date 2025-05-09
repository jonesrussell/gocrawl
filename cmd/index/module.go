// Package index provides commands for managing Elasticsearch index.
package index

import (
	"go.uber.org/fx"
)

// Module provides the index module for dependency injection.
var Module = fx.Module("index",
	fx.Provide(
		// Provide the index components
		NewCreator,
		NewLister,
		NewTableRenderer,
		NewDeleter,
	),
)
