package crawler

import (
	"fmt"
	"io"

	"github.com/jonesrussell/gocrawl/internal/common"
)

// Logger defines the interface required for colly debugging.
type Logger interface {
	Printf(format string, v ...any)
	Println(v ...any)
}

// DebugLogger implements both Logger and io.Writer interfaces for colly debugging.
type DebugLogger struct {
	logger common.Logger
}

// NewDebugLogger creates a new debug logger instance.
func NewDebugLogger(logger common.Logger) *DebugLogger {
	return &DebugLogger{
		logger: logger,
	}
}

// Write implements io.Writer interface.
func (d *DebugLogger) Write(p []byte) (int, error) {
	d.logger.Debug(string(p))
	return len(p), nil
}

// Printf implements Logger interface.
func (d *DebugLogger) Printf(format string, v ...any) {
	d.logger.Debug(fmt.Sprintf(format, v...))
}

// Println implements Logger interface.
func (d *DebugLogger) Println(v ...any) {
	d.logger.Debug(fmt.Sprint(v...))
}

// Ensure DebugLogger implements both interfaces
var (
	_ io.Writer = (*DebugLogger)(nil)
	_ Logger    = (*DebugLogger)(nil)
)
