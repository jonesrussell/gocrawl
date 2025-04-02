// Package sources provides the sources command implementation.
package sources

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Module provides the sources command module.
type Module struct {
	logger logger.Logger
}

// NewModule creates a new sources module.
func NewModule(log logger.Logger) *Module {
	return &Module{
		logger: log,
	}
}
