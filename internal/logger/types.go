// Package logger provides logging functionality for the application.
package logger

// Level represents the logging level.
type Level string

const (
	// DebugLevel logs debug messages.
	DebugLevel Level = "debug"
	// InfoLevel logs info messages.
	InfoLevel Level = "info"
	// WarnLevel logs warning messages.
	WarnLevel Level = "warn"
	// ErrorLevel logs error messages.
	ErrorLevel Level = "error"
	// FatalLevel logs fatal messages and exits.
	FatalLevel Level = "fatal"
)

// ANSI color codes for log levels
const (
	ColorDebug = "\033[36m" // Cyan
	ColorInfo  = "\033[32m" // Green
	ColorWarn  = "\033[33m" // Yellow
	ColorError = "\033[31m" // Red
	ColorFatal = "\033[35m" // Magenta
	ColorReset = "\033[0m"  // Reset
)

// Config represents the logger configuration.
type Config struct {
	// Level is the minimum logging level.
	Level Level `yaml:"level" json:"level"`
	// Development enables development mode.
	Development bool `yaml:"development" json:"development"`
	// Encoding sets the logger's encoding.
	Encoding string `yaml:"encoding" json:"encoding"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string `yaml:"outputPaths" json:"outputPaths"`
	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	ErrorOutputPaths []string `yaml:"errorOutputPaths" json:"errorOutputPaths"`
	// EnableColor enables colored output in development mode.
	EnableColor bool `yaml:"enableColor" json:"enableColor"`
}
