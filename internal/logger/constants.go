// Package logger provides logging functionality for the application.
package logger

// Default configuration values.
const (
	// DefaultLevel is the default logging level.
	DefaultLevel = InfoLevel
	// DefaultEncoding is the default log encoding format.
	DefaultEncoding = "console"
	// DefaultDevelopment is the default development mode setting.
	DefaultDevelopment = false
)

// Default output paths.
var (
	// DefaultOutputPaths is the default list of paths to write log output to.
	DefaultOutputPaths = []string{"stdout"}
	// DefaultErrorOutputPaths is the default list of paths to write internal logger errors to.
	DefaultErrorOutputPaths = []string{"stderr"}
)
