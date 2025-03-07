package common

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Type aliases for commonly used interfaces
type (
	Logger  = logger.Interface
	Storage = storage.Interface
	Config  = *config.Config
)

// Module provides shared dependencies for commands.
var Module = fx.Options(
	// Suppress fx logging
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),

	// Core modules used by most commands
	config.Module,  // Provides *config.Config
	logger.Module,  // Provides logger.Interface
	storage.Module, // Commonly used storage module
)
