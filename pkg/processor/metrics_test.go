package processor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector(t *testing.T) {
	t.Parallel()

	t.Run("new_collector", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		require.NotNil(t, collector)
		metrics := collector.GetMetrics()
		require.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.ProcessedCount)
		assert.Equal(t, int64(0), metrics.ErrorCount)
		assert.Equal(t, time.Duration(0), metrics.ProcessingDuration)
		assert.Equal(t, time.Time{}, metrics.LastProcessedTime)
	})

	t.Run("record_processing_time", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		duration := 100 * time.Millisecond
		collector.RecordProcessingTime(duration)
		metrics := collector.GetMetrics()
		assert.Equal(t, duration, metrics.ProcessingDuration)
	})

	t.Run("record_elements_processed", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		collector.RecordElementsProcessed(5)
		metrics := collector.GetMetrics()
		assert.Equal(t, int64(5), metrics.ProcessedCount)
	})

	t.Run("record_error", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		collector.RecordError()
		metrics := collector.GetMetrics()
		assert.Equal(t, int64(1), metrics.ErrorCount)
	})

	t.Run("reset", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		collector.RecordProcessingTime(100 * time.Millisecond)
		collector.RecordElementsProcessed(5)
		collector.RecordError()
		collector.Reset()
		metrics := collector.GetMetrics()
		assert.Equal(t, int64(0), metrics.ProcessedCount)
		assert.Equal(t, int64(0), metrics.ErrorCount)
		assert.Equal(t, time.Duration(0), metrics.ProcessingDuration)
		assert.Equal(t, time.Time{}, metrics.LastProcessedTime)
	})

	t.Run("concurrent_access", func(t *testing.T) {
		t.Parallel()
		collector := NewMetricsCollector()
		done := make(chan struct{})
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					collector.RecordElementsProcessed(1)
					collector.RecordError()
				}
				done <- struct{}{}
			}()
		}
		for i := 0; i < 10; i++ {
			<-done
		}
		metrics := collector.GetMetrics()
		assert.Equal(t, int64(1000), metrics.ProcessedCount)
		assert.Equal(t, int64(1000), metrics.ErrorCount)
	})
}
