package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"go.uber.org/zap"
)

type zapWrapper struct {
	logger *zap.Logger
}

func (w *zapWrapper) Debug(msg string, fields ...any) {
	w.logger.Debug(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Error(msg string, fields ...any) {
	w.logger.Error(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Info(msg string, fields ...any) {
	w.logger.Info(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Warn(msg string, fields ...any) {
	w.logger.Warn(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Fatal(msg string, fields ...any) {
	w.logger.Fatal(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Printf(format string, args ...any) {
	w.logger.Info(format, zap.Any("args", args))
}

func (w *zapWrapper) Errorf(format string, args ...any) {
	w.logger.Error(format, zap.Any("args", args))
}

func (w *zapWrapper) Sync() error {
	return w.logger.Sync()
}

// NewTestLogger creates a new test logger.
func NewTestLogger() common.Logger {
	logger, _ := zap.NewDevelopment()
	return &zapWrapper{logger: logger}
}
