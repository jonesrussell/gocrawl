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

func TestHTTPRequestMetrics(t *testing.T) {
	metrics := metrics.NewMetrics()

	// Test successful requests
	metrics.IncrementSuccessfulRequests()
	metrics.IncrementSuccessfulRequests()
	assert.Equal(t, int64(2), metrics.GetSuccessfulRequests(), "Should have 2 successful requests")

	// Test failed requests
	metrics.IncrementFailedRequests()
	assert.Equal(t, int64(1), metrics.GetFailedRequests(), "Should have 1 failed request")

	// Test rate limited requests
	metrics.IncrementRateLimitedRequests()
	metrics.IncrementRateLimitedRequests()
	assert.Equal(t, int64(2), metrics.GetRateLimitedRequests(), "Should have 2 rate limited requests")

	// Test reset
	metrics.ResetMetrics()
	assert.Equal(t, int64(0), metrics.GetSuccessfulRequests(), "Should have no successful requests after reset")
	assert.Equal(t, int64(0), metrics.GetFailedRequests(), "Should have no failed requests after reset")
	assert.Equal(t, int64(0), metrics.GetRateLimitedRequests(), "Should have no rate limited requests after reset")
}

func TestHTTPRequestMetricsConcurrently(t *testing.T) {
	metrics := metrics.NewMetrics()

	// Start goroutines to update metrics
	go func() {
		metrics.IncrementSuccessfulRequests()
	}()
	go func() {
		metrics.IncrementFailedRequests()
	}()
	go func() {
		metrics.IncrementRateLimitedRequests()
	}()

	// Wait for goroutines to complete
	time.Sleep(common.DefaultTestSleepDuration)

	// Verify metrics
	assert.Equal(t, int64(1), metrics.GetSuccessfulRequests(), "Should have 1 successful request")
	assert.Equal(t, int64(1), metrics.GetFailedRequests(), "Should have 1 failed request")
	assert.Equal(t, int64(1), metrics.GetRateLimitedRequests(), "Should have 1 rate limited request")
}
