// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	sourcestest "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the crawl module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := configtest.NewMockConfig().WithSources([]config.Source{
		{
			Name: "Test Source",
			URL:  "http://test.example.com",
		},
	})

	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "http://test.example.com",
			RateLimit: "1s",
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() *sources.Sources { return testSources },
		),
		fx.Replace(
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.As(new(config.Interface)),
			),
		),
		fx.Replace(
			fx.Annotate(
				func() api.IndexManager { return nil },
				fx.As(new(api.IndexManager)),
			),
		),
		crawl.Module,
	)

	require.NoError(t, app.Err())
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create test dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := configtest.NewMockConfig().
		WithSources([]config.Source{
			{
				Name: "Test Source",
				URL:  "https://test.com",
			},
		}).
		WithCrawlerConfig(&config.CrawlerConfig{
			MaxDepth:    3,
			Parallelism: 2,
		})

	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "https://test.com",
			RateLimit: "1s",
			MaxDepth:  2,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	var crawlerInstance crawler.Interface

	// Create test app with crawl module
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() *sources.Sources { return testSources },
			func() api.IndexManager { return api.NewMockIndexManager() },
		),
		// Provide only crawler module since we're providing sources directly
		crawler.Module,
		fx.Populate(&crawlerInstance),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify crawler configuration
	assert.NotNil(t, crawlerInstance, "Crawler should be provided")
	// Note: Add more specific crawler configuration checks here once crawler exposes them
}
