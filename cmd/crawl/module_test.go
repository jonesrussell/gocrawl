// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/testutils"
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
		fx.Replace(crawl.Module), // Use fx.Replace to override any existing providers
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
		fx.Replace(crawl.Module), // Use fx.Replace to override any existing providers
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
		fx.Replace(crawl.Module), // Use fx.Replace to override any existing providers
		fx.Provide(
			// Provide processors as a group
			fx.Annotate(
				func() []common.Processor {
					return []common.Processor{}
				},
				fx.ResultTags(`group:"processors"`),
			),
			// Provide named context
			fx.Annotate(
				func() context.Context {
					return context.Background()
				},
				fx.ResultTags(`name:"crawlContext"`),
			),
			// Provide source name
			fx.Annotate(
				func() string {
					return "test-source"
				},
				fx.ResultTags(`name:"sourceName"`),
			),
			// Provide signal handler
			fx.Annotate(
				func() *signal.SignalHandler {
					return &signal.SignalHandler{}
				},
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Provide article channel
			fx.Annotate(
				func() chan *models.Article {
					return make(chan *models.Article)
				},
				fx.ResultTags(`name:"crawlerArticleChannel"`),
			),
		),
		fx.Invoke(func(deps crawl.CommandDeps) {
			require.NotNil(t, deps.Context)
			require.NotNil(t, deps.Sources)
			require.NotNil(t, deps.Crawler)
			require.NotNil(t, deps.Logger)
			require.NotNil(t, deps.Config)
			require.NotNil(t, deps.Storage)
			require.NotNil(t, deps.Processors)
			require.NotEmpty(t, deps.SourceName)
			require.NotNil(t, deps.ArticleChan)
		}),
	)

	require.NoError(t, app.Err())
}
