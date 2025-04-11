package crawler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrawlerMetrics(t *testing.T) {
	t.Parallel()

	t.Run("NewCrawlerMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := NewCrawlerMetrics()
		require.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.GetProcessedCount())
		assert.Equal(t, int64(0), metrics.GetErrorCount())
		assert.True(t, metrics.GetLastProcessedTime().IsZero())
		assert.Equal(t, time.Duration(0), metrics.GetProcessingDuration())
	})

	t.Run("UpdateMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := NewCrawlerMetrics()

		// Update metrics
		startTime := time.Now()
		metrics.Update(startTime, 10, 2)

		// Verify metrics
		assert.Equal(t, int64(10), metrics.GetProcessedCount())
		assert.Equal(t, int64(2), metrics.GetErrorCount())
		assert.True(t, metrics.GetLastProcessedTime().After(startTime))
		assert.True(t, metrics.GetProcessingDuration() > 0)
	})

	t.Run("ResetMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := NewCrawlerMetrics()

		// Update metrics
		startTime := time.Now()
		metrics.Update(startTime, 10, 2)

		// Reset metrics
		metrics.Reset()

		// Verify metrics are reset
		assert.Equal(t, int64(0), metrics.GetProcessedCount())
		assert.Equal(t, int64(0), metrics.GetErrorCount())
		assert.True(t, metrics.GetLastProcessedTime().IsZero())
		assert.Equal(t, time.Duration(0), metrics.GetProcessingDuration())
	})

	t.Run("ConcurrentUpdates", func(t *testing.T) {
		t.Parallel()
		metrics := NewCrawlerMetrics()
		startTime := time.Now()

		// Start multiple goroutines to update metrics
		for i := 0; i < 100; i++ {
			go func() {
				metrics.Update(startTime, 1, 0)
			}()
		}

		// Wait for all goroutines to complete
		time.Sleep(100 * time.Millisecond)

		// Verify metrics
		assert.Equal(t, int64(100), metrics.GetProcessedCount())
		assert.True(t, metrics.GetLastProcessedTime().After(startTime))
		assert.True(t, metrics.GetProcessingDuration() > 0)
	})
}
