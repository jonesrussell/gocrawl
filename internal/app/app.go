package app

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/zap"
)

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config) (logger.Interface, error) {
	return logger.NewCustomLogger(logger.Params{
		Debug:  cfg.App.Debug,
		Level:  zap.InfoLevel,
		AppEnv: cfg.App.Environment,
	})
}
