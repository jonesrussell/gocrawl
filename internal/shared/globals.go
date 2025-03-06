package shared

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

var (
	// Config is the application configuration
	Config *config.Config

	// Logger is the application logger
	Logger logger.Interface
)

// SetConfig sets the global configuration
func SetConfig(cfg *config.Config) {
	Config = cfg
}

// SetLogger sets the global logger
func SetLogger(l logger.Interface) {
	Logger = l
}
