// Package sources implements the command-line interface for managing content sources.
package sources

import (
	"go.uber.org/fx"
)

// Module provides the sources command dependencies
var Module = fx.Module("sourcesCmd",
	fx.Provide(Command),
)
