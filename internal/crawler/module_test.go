// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// setupTestApp creates a new test application with all required dependencies.
// It provides mock implementations of all interfaces required by the crawler module.
func setupTestApp(t *testing.T) *fxtest.App {
	t.Helper()

	return fxtest.New(t,
		fx.NopLogger,
		// Provide core dependencies
		fx.Provide(
			// Named dependencies
			fx.Annotate(
				func() context.Context { return t.Context() },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() config.Interface { return &mockConfig{} },
				fx.ResultTags(`name:"config"`),
			),
			fx.Annotate(
				func() sources.Interface { return &mockSources{} },
				fx.ResultTags(`name:"sourceManager"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return signal.NewSignalHandler(nil) },
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Logger provider that replaces the default logger.Module provider
			fx.Annotate(
				func() logger.Interface { return logger.NewNoOp() },
				fx.ResultTags(`name:"logger"`),
				fx.As(new(logger.Interface)),
			),
			// Provide unnamed interfaces required by the crawler module
			func() types.Interface { return &mockStorage{} },
			func() api.IndexManager { return &mockIndexManager{} },
			func() api.SearchManager { return &mockSearchManager{} },
			func() config.Interface { return &mockConfig{} },
		),
		// Supply done channel
		fx.Supply(
			fx.Annotate(
				make(chan struct{}),
				fx.ResultTags(`name:"done"`),
			),
		),
		// Mock content processors
		fx.Provide(
			fx.Annotate(
				func() models.ContentProcessor {
					return &mockContentProcessor{}
				},
				fx.ResultTags(`group:"processors"`),
			),
		),
		// Include only required modules
		sources.Module,
		crawler.Module,
		// Verify dependencies
		fx.Invoke(func(deps crawler.CrawlDeps) {
			verifyDependencies(t, &deps)
		}),
	)
}

// verifyDependencies checks that all required dependencies are present in the CrawlDeps struct.
func verifyDependencies(t *testing.T, deps *crawler.CrawlDeps) {
	t.Helper()

	require.NotNil(t, deps)
	require.NotNil(t, deps.Config)
	require.NotNil(t, deps.Storage)
	require.NotNil(t, deps.Sources)
	require.NotNil(t, deps.ArticleChan)
}

// TestDependencyInjection verifies that all dependencies are properly injected into the CrawlDeps struct.
func TestDependencyInjection(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestModuleConstruction verifies that the crawler module can be constructed with all required dependencies.
func TestModuleConstruction(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestModuleLifecycle verifies that the crawler module can be started and stopped correctly.
func TestModuleLifecycle(t *testing.T) {
	app := setupTestApp(t)
	require.NoError(t, app.Start(t.Context()))
	require.NoError(t, app.Stop(t.Context()))
}
