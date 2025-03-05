package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockSource represents a simplified version of sources.Source for testing
type mockSource struct {
	Time []string
}

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
			source := &mockSource{
				Time: tt.scheduledTimes,
			}

			shouldRunCount := 0
			for _, scheduledTime := range source.Time {
				parsedTime, err := time.Parse("15:04", scheduledTime)
				assert.NoError(t, err)

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

			assert.NoError(t, err)
			assert.Equal(t, tt.expectHour, parsedTime.Hour())
		})
	}
}
