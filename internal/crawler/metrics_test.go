package crawler_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	t.Run("NewMetrics", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics()
		assert.NotNil(t, m)
		assert.NotZero(t, m.GetStartTime())
	})

	t.Run("IncrementProcessed", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics().(*crawler.Metrics)

		// Test initial count
		assert.Equal(t, int64(0), m.GetProcessedCount())

		// Test increment
		m.IncrementProcessed()
		assert.Equal(t, int64(1), m.GetProcessedCount())

		// Test multiple increments
		for range 5 {
			m.IncrementProcessed()
		}
		assert.Equal(t, int64(6), m.GetProcessedCount())
	})

	t.Run("IncrementError", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics().(*crawler.Metrics)

		// Test initial count
		assert.Equal(t, int64(0), m.GetErrorCount())

		// Test increment
		m.IncrementError()
		assert.Equal(t, int64(1), m.GetErrorCount())

		// Test multiple increments
		for range 5 {
			m.IncrementError()
		}
		assert.Equal(t, int64(6), m.GetErrorCount())
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics().(*crawler.Metrics)

		// Start multiple goroutines to access metrics
		done := make(chan struct{})
		for range 10 {
			go func() {
				for range 100 {
					m.IncrementProcessed()
					m.IncrementError()
					m.GetProcessedCount()
					m.GetErrorCount()
					m.GetStartTime()
				}
				done <- struct{}{}
			}()
		}

		// Wait for all goroutines to complete
		for range 10 {
			<-done
		}

		// Verify final counts
		assert.Equal(t, int64(1000), m.GetProcessedCount())
		assert.Equal(t, int64(1000), m.GetErrorCount())
	})

	t.Run("Reset", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics().(*crawler.Metrics)

		// Add some metrics
		for range 5 {
			m.IncrementProcessed()
			m.IncrementError()
		}

		// Store original start time
		originalStartTime := m.GetStartTime()

		// Reset metrics
		m.Reset()

		// Verify reset
		assert.Equal(t, int64(0), m.GetProcessedCount())
		assert.Equal(t, int64(0), m.GetErrorCount())
		assert.True(t, m.GetStartTime().After(originalStartTime))
	})

	t.Run("StartTime", func(t *testing.T) {
		t.Parallel()
		m := crawler.NewMetrics().(*crawler.Metrics)

		// Verify start time is recent
		startTime := m.GetStartTime()
		assert.Less(t, time.Since(startTime), time.Second)

		// Verify start time doesn't change
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, startTime, m.GetStartTime())
	})
}
