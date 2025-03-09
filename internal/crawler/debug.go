package crawler

import (
	"github.com/jonesrussell/gocrawl/internal/common"
)

// debugLogger adapts common.Logger to io.Writer for colly debugging.
type debugLogger struct {
	logger common.Logger
}

// Write implements io.Writer.
func (d *debugLogger) Write(p []byte) (int, error) {
	d.logger.Debug(string(p))
	return len(p), nil
}

// newDebugLogger creates a new debug logger adapter.
func newDebugLogger(logger common.Logger) *debugLogger {
	return &debugLogger{logger: logger}
}
