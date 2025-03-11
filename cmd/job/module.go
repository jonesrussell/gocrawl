// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Module provides the job scheduler command dependencies
var Module = fx.Module("job",
	sources.Module,
)
