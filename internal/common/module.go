// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This file specifically handles
// dependency injection and module configuration using the fx framework.
package common

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Type aliases for commonly used interfaces and types.
// These aliases provide a convenient way to reference core types
// throughout the application while maintaining clear dependencies.
type (
	// Storage is an alias for the storage interface, providing
	// data persistence operations across the application.
	Storage = storage.Interface

	// Config is an alias for the configuration interface, providing
	// access to application-wide settings.
	Config = config.Interface

	// Logger is an alias for the logger interface, providing
	// structured logging capabilities across the application.
	Logger = logger.Interface
)

// Module provides shared dependencies for commands.
// It combines core modules that are commonly used across
// different parts of the application, setting up the
// dependency injection framework and core services.
var Module = fx.Module("common",
	// Suppress fx logging to reduce noise in the application logs.
	// This replaces the default fx logger with a no-op implementation.
	fx.WithLogger(func() fxevent.Logger {
		return &fxevent.NopLogger
	}),

	// Core modules used by most commands.
	config.Module,
	logger.Module,
	storage.Module,
)
