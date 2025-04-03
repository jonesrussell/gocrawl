// Package test provides test utilities for the indices command.
package test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// TestModule provides a test module with mock dependencies
func TestModule(t *testing.T) fx.Option {
	return fx.Module("test",
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface { return configtestutils.NewMockConfig() },
			logger.NewNoOp,
		),
	)
}
