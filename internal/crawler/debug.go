// Package crawler provides web crawling functionality.
package crawler

import (
	"fmt"
	"io"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// DebugLogger is a wrapper around the main logger for debug-specific logging.
type DebugLogger struct {
	logger.Logger
}

// NewDebugLogger creates a new debug logger.
func NewDebugLogger(log logger.Logger) *DebugLogger {
	return &DebugLogger{Logger: log}
}

// Write implements io.Writer interface.
func (d *DebugLogger) Write(p []byte) (int, error) {
	d.Debug(string(p))
	return len(p), nil
}

// Printf implements Logger interface.
func (d *DebugLogger) Printf(format string, v ...any) {
	d.Debug(fmt.Sprintf(format, v...))
}

// Println implements Logger interface.
func (d *DebugLogger) Println(v ...any) {
	d.Debug(fmt.Sprint(v...))
}

// Ensure DebugLogger implements both interfaces
var (
	_ io.Writer     = (*DebugLogger)(nil)
	_ logger.Logger = (*DebugLogger)(nil)
)
