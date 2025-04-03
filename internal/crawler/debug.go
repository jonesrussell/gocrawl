// Package crawler provides web crawling functionality.
package crawler

import (
	"fmt"
	"io"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// DebugLogger is a wrapper around the main logger for debug-specific logging.
type DebugLogger struct {
	log logger.Interface
}

// NewDebugLogger creates a new debug logger.
func NewDebugLogger(log logger.Interface) *DebugLogger {
	return &DebugLogger{log: log}
}

// Write implements io.Writer interface.
func (d *DebugLogger) Write(p []byte) (int, error) {
	d.log.Debug(string(p))
	return len(p), nil
}

// Printf implements Logger interface.
func (d *DebugLogger) Printf(format string, v ...any) {
	d.log.Debug(fmt.Sprintf(format, v...))
}

// Println implements Logger interface.
func (d *DebugLogger) Println(v ...any) {
	d.log.Debug(fmt.Sprint(v...))
}

// Debug logs a debug message.
func (d *DebugLogger) Debug(msg string, fields ...any) {
	d.log.Debug(msg, fields...)
}

// Info logs an info message.
func (d *DebugLogger) Info(msg string, fields ...any) {
	d.log.Info(msg, fields...)
}

// Warn logs a warning message.
func (d *DebugLogger) Warn(msg string, fields ...any) {
	d.log.Warn(msg, fields...)
}

// Error logs an error message.
func (d *DebugLogger) Error(msg string, fields ...any) {
	d.log.Error(msg, fields...)
}

// Fatal logs a fatal message and exits.
func (d *DebugLogger) Fatal(msg string, fields ...any) {
	d.log.Fatal(msg, fields...)
}

// With creates a child logger with additional fields.
func (d *DebugLogger) With(fields ...any) logger.Interface {
	return &DebugLogger{log: d.log.With(fields...)}
}

// Ensure DebugLogger implements both interfaces
var (
	_ io.Writer        = (*DebugLogger)(nil)
	_ logger.Interface = (*DebugLogger)(nil)
)
