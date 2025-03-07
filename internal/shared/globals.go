// Package shared provides global variables and functions that need to be
// accessed across different parts of the GoCrawl application. While global
// state should generally be avoided, this package provides a controlled way
// to share essential application-wide resources.
package shared

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Global variables for application-wide resources.
// These variables should be set early in the application lifecycle
// and should remain constant after initialization.
var (
	// Config is the global application configuration.
	// It provides access to all configuration settings and should be
	// set during application startup using SetConfig.
	Config *config.Config

	// Logger is the global application logger.
	// It provides structured logging capabilities and should be
	// set during application startup using SetLogger.
	Logger logger.Interface
)

// SetConfig sets the global configuration instance.
// This function should be called during application initialization
// before any other components attempt to access the configuration.
//
// Parameters:
//   - cfg: The configuration instance to set globally
func SetConfig(cfg *config.Config) {
	Config = cfg
}

// SetLogger sets the global logger instance.
// This function should be called during application initialization
// before any other components attempt to use logging capabilities.
//
// Parameters:
//   - l: The logger interface implementation to set globally
func SetLogger(l logger.Interface) {
	Logger = l
}
