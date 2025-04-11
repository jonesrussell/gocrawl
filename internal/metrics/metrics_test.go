package metrics_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	metrics := metrics.NewMetrics()
	assert.NotNil(t, metrics)
	assert.False(t, metrics.GetStartTime().IsZero())
}

func TestUpdateMetrics(t *testing.T) {
	metrics := metrics.NewMetrics()

	// Test successful processing
	metrics.UpdateMetrics(true)
	assert.Equal(t, int64(1), metrics.GetProcessedCount())
	assert.Equal(t, int64(0), metrics.GetErrorCount())
	assert.False(t, metrics.GetLastProcessedTime().IsZero())

	// Test error processing
	metrics.UpdateMetrics(false)
	assert.Equal(t, int64(2), metrics.GetProcessedCount())
	assert.Equal(t, int64(1), metrics.GetErrorCount())
}

func TestResetMetrics(t *testing.T) {
	metrics := metrics.NewMetrics()
	metrics.UpdateMetrics(true)
	metrics.UpdateMetrics(false)
	metrics.SetCurrentSource("test")

	metrics.ResetMetrics()

	assert.Equal(t, int64(0), metrics.GetProcessedCount())
	assert.Equal(t, int64(0), metrics.GetErrorCount())
	assert.True(t, metrics.GetLastProcessedTime().IsZero())
	assert.Empty(t, metrics.GetCurrentSource())
}

func TestProcessingDuration(t *testing.T) {
	metrics := metrics.NewMetrics()
	initialDuration := metrics.GetProcessingDuration()

	time.Sleep(100 * time.Millisecond)
	updatedDuration := metrics.GetProcessingDuration()

	assert.Greater(t, updatedDuration, initialDuration)
}

func TestCurrentSource(t *testing.T) {
	metrics := metrics.NewMetrics()
	assert.Empty(t, metrics.GetCurrentSource())

	metrics.SetCurrentSource("test")
	assert.Equal(t, "test", metrics.GetCurrentSource())
}

func TestUpdateMetricsConcurrently(t *testing.T) {
	metrics := metrics.NewMetrics()

	// Start a goroutine to update metrics
	go func() {
		metrics.UpdateMetrics(true)
	}()

	// Wait for goroutine to complete
	time.Sleep(common.DefaultTestSleepDuration)

	// Verify metrics
	assert.Equal(t, int64(1), metrics.GetProcessedCount())
	assert.Equal(t, int64(0), metrics.GetErrorCount())
}
