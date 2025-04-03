// Package sources provides the sources command implementation.
package sources

import (
	"go.uber.org/fx"
)

// Module provides the sources command functionality.
var Module = fx.Module("sources",
	fx.Provide(NewSourcesCommand),
)
