// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This file specifically handles
// dependency injection and module configuration using the fx framework.
package common

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
	sources.Module,
)
