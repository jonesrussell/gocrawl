// Package crawl_test implements tests for the crawl command module.
package crawl_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the crawl module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockCfg := config.NewMockConfig().WithSources([]config.Source{
		{
			Name: "Test Source",
			URL:  "https://test.com",
		},
	})

	testSources := &sources.Sources{
		Sources: []sources.Config{
			{
				Name:      "Test Source",
				URL:       "https://test.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
		},
	}

	var (
		crawlerInstance *crawler.Crawler
		src             *sources.Sources
	)

	// Create test app with crawl module
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() *sources.Sources { return testSources },
		),
		// Provide only crawler module since we're providing sources directly
		crawler.Module,
		fx.Populate(&crawlerInstance, &src),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify dependencies were provided
	assert.NotNil(t, crawlerInstance, "Crawler should be provided")
	assert.NotNil(t, src, "Sources should be provided")
	assert.Equal(t, "Test Source", src.Sources[0].Name, "Source name should match configuration")
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockCfg := config.NewMockConfig().
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

	testSources := &sources.Sources{
		Sources: []sources.Config{
			{
				Name:      "Test Source",
				URL:       "https://test.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
		},
	}

	var crawlerInstance *crawler.Crawler

	// Create test app with crawl module
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() *sources.Sources { return testSources },
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
