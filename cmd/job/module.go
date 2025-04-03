// Package job provides the job command implementation.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Module provides the job command module.
type Module struct {
	logger logger.Interface
}

// NewModule creates a new job module.
func NewModule(log logger.Interface) *Module {
	return &Module{
		logger: log,
	}
}
