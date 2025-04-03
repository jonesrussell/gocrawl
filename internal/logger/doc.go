// Package logger provides a structured logging interface for the application.
// It is built on top of zap and provides a simple interface for logging
// operations with support for structured logging, log levels, and field injection.
//
// The package exports the following types:
//
//	Interface - The main logging interface
//	Config   - Configuration for the logger
//	Level    - Log level type
//
// Example usage:
//
//	// Create a new logger with default configuration
//	log, err := logger.Constructor(logger.Params{})
//	if err != nil {
//	    panic(err)
//	}
//	defer log.Sync()
//
//	// Log messages with fields
//	log.Info("User logged in", "user_id", 123, "ip", "192.168.1.1")
//
//	// Create a child logger with additional fields
//	userLog := log.With("user_id", 123)
//	userLog.Info("User action", "action", "view_profile")
//
//	// Log at different levels
//	log.Debug("Debug message", "key", "value")
//	log.Info("Info message", "key", "value")
//	log.Warn("Warning message", "key", "value")
//	log.Error("Error message", "key", "value")
//	log.Fatal("Fatal message", "key", "value")
//
// Configuration:
//
//	config := &logger.Config{
//	    Level:       logger.InfoLevel,
//	    Development: false,
//	    Encoding:    "json",
//	    OutputPaths: []string{"stdout"},
//	}
//	log, err := logger.Constructor(logger.Params{Config: config})
//
// Dependency Injection:
//
//	// In your fx application
//	app := fx.New(
//	    logger.Module,
//	    // other modules...
//	)
//
// The logger package is designed to be used with dependency injection
// and provides a clean interface for logging operations. It supports
// structured logging with fields, different log levels, and child loggers
// with additional fields.
package logger
