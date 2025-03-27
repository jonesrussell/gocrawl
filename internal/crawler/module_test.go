// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"testing"

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

// TestConfigModule provides a test-specific config module that doesn't try to load files.
var TestConfigModule = fx.Module("testConfig",
	fx.Provide(
		fx.Annotate(
			func() config.Interface { return &mockConfig{} },
			fx.ResultTags(`name:"config"`),
		),
	),
)

// setupTestApp creates a new test application with all required dependencies.
// It provides mock implementations of all interfaces required by the crawler module.
func setupTestApp(t *testing.T) *fxtest.App {
	t.Helper()

	// Set environment variables to prevent file loading
	t.Setenv("APP_ENV", "test")
	t.Setenv("CONFIG_PATH", "")

	return fxtest.New(t,
		fx.NopLogger,
		// Provide core dependencies
		fx.Provide(
			// Named dependencies
			fx.Annotate(
				func() sources.Interface { return &mockSources{} },
				fx.ResultTags(`name:"testSourceManager"`),
			),
			// Logger provider that replaces the default logger.Module provider
			fx.Annotate(
				logger.NewNoOp,
				fx.ResultTags(`name:"logger"`),
				fx.As(new(logger.Interface)),
			),
			// Provide unnamed interfaces required by the crawler module
			func() types.Interface { return &mockStorage{} },
			fx.Annotate(
				func() api.IndexManager { return &mockIndexManager{} },
				fx.ResultTags(`name:"indexManager"`),
			),
			func() api.SearchManager { return &mockSearchManager{} },
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
		// Include test config module and crawler module
		TestConfigModule,
		crawler.Module,
		// Verify dependencies
		fx.Invoke(func(p crawler.Params) {
			verifyDependencies(t, &p)
		}),
	)
}

// verifyDependencies checks that all required dependencies are present in the Params struct.
func verifyDependencies(t *testing.T, p *crawler.Params) {
	t.Helper()

	require.NotNil(t, p)
	require.NotNil(t, p.Logger)
	require.NotNil(t, p.Sources)
	require.NotNil(t, p.IndexManager)
}

// TestDependencyInjection verifies that all dependencies are properly injected into the Params struct.
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
