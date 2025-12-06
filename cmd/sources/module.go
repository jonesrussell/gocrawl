// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/cmd/common"
	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
// Note: Command registration is handled by NewSourcesCommand() function, not through FX Group annotation.
var Module = fx.Module("cmd_sources",
	// Include required modules
	common.Module,

	// Provide command dependencies
	fx.Provide(
		NewTableRenderer,
		NewLister,
	),
)
