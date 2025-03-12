// Package job_test implements tests for the job scheduler command.
package job_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestJobScheduling tests the job scheduling functionality
func TestJobScheduling(t *testing.T) {
	tests := []struct {
		name           string
		currentTime    time.Time
		scheduledTimes []string
		shouldRun      bool
	}{
		{
			name:           "AM scheduled time matches",
			currentTime:    time.Date(2024, 3, 5, 3, 13, 0, 0, time.UTC),
			scheduledTimes: []string{"03:13", "15:13"},
			shouldRun:      true,
		},
		{
			name:           "PM scheduled time matches",
			currentTime:    time.Date(2024, 3, 5, 15, 13, 0, 0, time.UTC),
			scheduledTimes: []string{"03:13", "15:13"},
			shouldRun:      true,
		},
		{
			name:           "No time match",
			currentTime:    time.Date(2024, 3, 5, 4, 13, 0, 0, time.UTC),
			scheduledTimes: []string{"03:13", "15:13"},
			shouldRun:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &sources.Config{
				Time: tt.scheduledTimes,
			}

			shouldRunCount := 0
			for _, scheduledTime := range source.Time {
				parsedTime, err := time.Parse("15:04", scheduledTime)
				require.NoError(t, err)

				if tt.currentTime.Hour() == parsedTime.Hour() && tt.currentTime.Minute() == parsedTime.Minute() {
					shouldRunCount++
				}
			}

			if tt.shouldRun {
				assert.Equal(t, 1, shouldRunCount, "Job should run exactly once when time matches")
			} else {
				assert.Equal(t, 0, shouldRunCount, "Job should not run when time doesn't match")
			}
		})
	}
}

// TestTimeFormatParsing tests the time format parsing functionality
func TestTimeFormatParsing(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		expectHour  int
		expectError bool
	}{
		{
			name:        "Early morning time",
			timeStr:     "03:13",
			expectHour:  3,
			expectError: false,
		},
		{
			name:        "Afternoon time",
			timeStr:     "15:13",
			expectHour:  15,
			expectError: false,
		},
		{
			name:        "Invalid time format",
			timeStr:     "3:13:00",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedTime, err := time.Parse("15:04", tt.timeStr)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectHour, parsedTime.Hour())
		})
	}
}

// TestJobCommand tests the job command functionality
func TestJobCommand(t *testing.T) {
	// Create a test logger
	mockLogger := &mockLogger{}

	// Create test sources
	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			Time:      []string{"03:13", "15:13"},
			URL:       "https://test.com",
			RateLimit: "1s",
			MaxDepth:  1,
		},
	}
	testSources := &sources.Sources{}
	testSources.SetSources(testConfigs)

	// Create mock config
	mockCfg := &mockConfig{}

	// Create a test command
	cmd := job.Command()

	// Create a test app with all necessary dependencies
	app := fxtest.New(t,
		fx.Provide(
			// Override the sources provider
			fx.Annotate(
				func() sources.Interface {
					return testSources
				},
				fx.ResultTags(`name:"sourceManager"`),
			),
			// Mock dependencies
			func() logger.Interface { return mockLogger },
			func() config.Interface { return mockCfg },
			func() crawler.Interface { return &mockCrawler{} },
		),
		job.Module,
		fx.Invoke(func(p job.Params) {
			startJobScheduler := func(cmd *cobra.Command, _ []string) error {
				rootPath := cmd.Root().Name()
				p.Logger.Info("Starting job scheduler", "root", rootPath)
				return nil
			}
			cmd.RunE = startJobScheduler
		}),
	)
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Test command
	require.NotNil(t, cmd)
	require.Equal(t, "job", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.NotEmpty(t, cmd.Long)

	// Test command execution
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.NoError(t, err)
}
