package crawler_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	t.Run("NewMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := crawler.NewCrawlerMetrics()
		require.NotNil(t, metrics)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		metrics := crawler.NewCrawlerMetrics()
		startTime := time.Now()
		metrics.Update(startTime, 10, 2)
		assert.Equal(t, int64(10), metrics.GetProcessedCount())
		assert.Equal(t, int64(2), metrics.GetErrorCount())
		assert.Positive(t, metrics.GetProcessingDuration())
	})

	t.Run("Reset", func(t *testing.T) {
		t.Parallel()
		metrics := crawler.NewCrawlerMetrics()
		startTime := time.Now()
		metrics.Update(startTime, 10, 2)
		metrics.Reset()
		assert.Zero(t, metrics.GetProcessedCount())
		assert.Zero(t, metrics.GetErrorCount())
		assert.Zero(t, metrics.GetProcessingDuration())
	})

	t.Run("GetProcessingDuration", func(t *testing.T) {
		t.Parallel()
		metrics := crawler.NewCrawlerMetrics()
		startTime := time.Now()
		metrics.Update(startTime, 10, 2)
		assert.Positive(t, metrics.GetProcessingDuration())
	})
}
