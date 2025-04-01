package common

import (
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/zap"
)

// NewZapLogger creates a new logger using zap.
//
// Parameters:
//   - zapLogger: The zap logger to wrap
//
// Returns:
//   - types.Logger: The wrapped logger
func NewZapLogger(zapLogger *zap.Logger) types.Logger {
	return &logger.ZapLogger{Logger: zapLogger}
}

// NewTestLogger creates a new logger for testing.
//
// Returns:
//   - types.Logger: A new test logger
func NewTestLogger() types.Logger {
	return logger.NewTestLogger()
}
