// Package job_test implements tests for the job scheduler command.
package job_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the job module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockCfg := config.NewMockConfig().WithSources([]config.Source{
		{
			Name: "Test Source",
			Time: []string{"03:13", "15:13"},
			URL:  "https://test.com",
		},
	})

	var (
		params job.Params
		src    *sources.Sources
	)

	testSources := &sources.Sources{
		Sources: []sources.Config{
			{
				Name:      "Test Source",
				Time:      []string{"03:13", "15:13"},
				URL:       "https://test.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
		},
	}

	// Create test app with job module
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() *sources.Sources { return testSources },
		),
		// Use only the components we need from job.Module
		fx.Provide(job.Command),
		fx.Populate(&params, &src),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify dependencies were provided
	assert.NotNil(t, params.Logger, "Logger should be injected into params")
	assert.NotNil(t, params.Sources, "Sources should be injected into params")
	assert.NotNil(t, src, "Sources should be provided")
	assert.Equal(t, "Test Source", src.Sources[0].Name, "Source name should match configuration")
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	// Create test dependencies with specific configuration
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "Starting job scheduler", "root", "job").Return()

	mockCfg := config.NewMockConfig().WithSources([]config.Source{
		{
			Name: "Test Source",
			Time: []string{"03:13", "15:13"},
			URL:  "https://test.com",
		},
	})

	testSources := &sources.Sources{
		Sources: []sources.Config{
			{
				Name:      "Test Source",
				Time:      []string{"03:13", "15:13"},
				URL:       "https://test.com",
				RateLimit: "1s",
				MaxDepth:  2,
			},
		},
	}

	var params job.Params

	// Create test app with job module
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() *sources.Sources { return testSources },
		),
		// Use only the components we need from job.Module
		fx.Provide(job.Command),
		fx.Populate(&params),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify configuration
	assert.Equal(t, mockLogger, params.Logger, "Logger should match provided mock")
	assert.NotNil(t, params.Sources, "Sources should be configured")
	assert.Len(t, params.Sources.Sources, 1, "Should have one source configured")
	assert.Equal(t, "Test Source", params.Sources.Sources[0].Name, "Source name should match configuration")
}
