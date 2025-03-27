package content

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the content service module
var Module = fx.Module("content",
	fx.Provide(
		NewService,
		NewProcessor,
	),
)

// Params defines the parameters for the content service
type Params struct {
	fx.In

	Logger    common.Logger
	Config    *config.Config
	Storage   types.Interface
	IndexName string
}
