// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

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
