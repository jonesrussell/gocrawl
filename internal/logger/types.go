// Package logger provides logging functionality for the application.
package logger

import "go.uber.org/zap"

// Level represents a logging level.
type Level string

// Available logging levels.
const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
)

// Config holds configuration for the logger.
type Config struct {
	// Level is the minimum level to log.
	Level Level `yaml:"level" json:"level"`
	// Development enables development mode.
	Development bool `yaml:"development" json:"development"`
	// Encoding is the log encoding format (json or console).
	Encoding string `yaml:"encoding" json:"encoding"`
	// OutputPaths is a list of paths to write log output to.
	OutputPaths []string `yaml:"output_paths" json:"output_paths"`
	// ErrorOutputPaths is a list of paths to write internal logger errors to.
	ErrorOutputPaths []string `yaml:"error_output_paths" json:"error_output_paths"`
}

// Logger wraps a zap logger with additional functionality.
type logger struct {
	*zap.Logger
	config *Config
}

// Ensure logger implements Interface
var _ Interface = (*logger)(nil)
