// Package indices implements the command-line interface for managing Elasticsearch indices.
package indices

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"go.uber.org/fx"
)

// Module provides the indices command dependencies
var Module = fx.Module("indices",
	api.Module,
)
