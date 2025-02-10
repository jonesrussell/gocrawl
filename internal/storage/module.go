package storage

import (
	"go.uber.org/fx"
)

// Module provides the storage as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("storage",
	fx.Provide(
		NewStorage,
	),
)
