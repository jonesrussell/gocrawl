// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
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
		fx.Supply(make(chan struct{})),
		fx.Provide(
			fx.Annotate(
				func() config.Interface { return &mockConfig{} },
				fx.ResultTags(`name:"config"`),
			),
			fx.Annotate(
				func() api.SearchManager { return &mockSearchManager{} },
				fx.ResultTags(`name:"searchManager"`),
			),
			fx.Annotate(
				func() api.IndexManager { return &mockIndexManager{} },
				fx.ResultTags(`name:"indexManager"`),
			),
			fx.Annotate(
				func() types.Interface { return &mockStorage{} },
				fx.ResultTags(`name:"storage"`),
			),
			fx.Annotate(
				func() sources.Interface { return &mockSources{} },
				fx.ResultTags(`name:"sources"`),
			),
		),
		crawler.Module,
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
