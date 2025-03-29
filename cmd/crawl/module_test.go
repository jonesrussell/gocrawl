// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/common/testutils"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the crawl module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	// Create test module with default configuration
	testModule := testutils.NewCommandTestModule(t)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule.Module(),
		crawl.Module,
		fx.Provide(
			fx.Annotate(
				configtestutils.NewMockConfig,
				fx.As(new(config.Interface)),
			),
		),
	)

	require.NoError(t, app.Err())
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	// Create test module with default configuration
	testModule := testutils.NewCommandTestModule(t)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule.Module(),
		crawl.Module,
		fx.Provide(
			fx.Annotate(
				configtestutils.NewMockConfig,
				fx.As(new(config.Interface)),
			),
		),
	)

	require.NoError(t, app.Err())
}

// TestCommandDeps tests that the command dependencies are properly injected
func TestCommandDeps(t *testing.T) {
	// Create test module with default configuration
	testModule := testutils.NewCommandTestModule(t)

	app := fxtest.New(t,
		fx.NopLogger,
		testModule.Module(),
		crawl.Module,
		fx.Provide(
			fx.Annotate(
				configtestutils.NewMockConfig,
				fx.As(new(config.Interface)),
			),
		),
		fx.Invoke(func(deps crawl.CommandDeps) {
			require.NotNil(t, deps.Lifecycle)
			require.NotNil(t, deps.Sources)
			require.NotNil(t, deps.Crawler)
			require.NotNil(t, deps.Logger)
			require.NotNil(t, deps.Config)
			require.NotNil(t, deps.Storage)
			require.NotNil(t, deps.Done)
			require.NotNil(t, deps.Context)
			require.NotNil(t, deps.Processors)
			require.NotEmpty(t, deps.SourceName)
			require.NotNil(t, deps.ArticleChan)
			require.NotNil(t, deps.Handler)
		}),
	)

	require.NoError(t, app.Err())
}
