package metrics_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/constants"
	"github.com/jonesrussell/gocrawl/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := metrics.NewMetrics()
	assert.NotNil(t, m)
	assert.False(t, m.GetStartTime().IsZero())
}

func TestUpdateMetrics(t *testing.T) {
	m := metrics.NewMetrics()

	// Test successful processing
	m.UpdateMetrics(true)
	assert.Equal(t, int64(1), m.GetProcessedCount())
	assert.Equal(t, int64(0), m.GetErrorCount())
	assert.False(t, m.GetLastProcessedTime().IsZero())

	// Test error processing
	m.UpdateMetrics(false)
	assert.Equal(t, int64(2), m.GetProcessedCount())
	assert.Equal(t, int64(1), m.GetErrorCount())
}

func TestResetMetrics(t *testing.T) {
	m := metrics.NewMetrics()
	m.UpdateMetrics(true)
	m.UpdateMetrics(false)
	m.SetCurrentSource("test")

	m.ResetMetrics()

	assert.Equal(t, int64(0), m.GetProcessedCount())
	assert.Equal(t, int64(0), m.GetErrorCount())
	assert.True(t, m.GetLastProcessedTime().IsZero())
	assert.Empty(t, m.GetCurrentSource())
}

func TestCurrentSource(t *testing.T) {
	m := metrics.NewMetrics()
	assert.Empty(t, m.GetCurrentSource())

	m.SetCurrentSource("test")
	assert.Equal(t, "test", m.GetCurrentSource())
}

func TestUpdateMetricsConcurrently(t *testing.T) {
	m := metrics.NewMetrics()

	// Start a goroutine to update metrics
	go func() {
		m.UpdateMetrics(true)
	}()

	// Wait for goroutine to complete
	time.Sleep(constants.DefaultTestSleepDuration)

	// Verify metrics
	assert.Equal(t, int64(1), m.GetProcessedCount())
	assert.Equal(t, int64(0), m.GetErrorCount())
}

func TestHTTPRequestMetrics(t *testing.T) {
	m := metrics.NewMetrics()

	// Test successful requests
	m.IncrementSuccessfulRequests()
	m.IncrementSuccessfulRequests()
	assert.Equal(t, int64(2), m.GetSuccessfulRequests(), "Should have 2 successful requests")

	// Test failed requests
	m.IncrementFailedRequests()
	assert.Equal(t, int64(1), m.GetFailedRequests(), "Should have 1 failed request")

	// Test rate limited requests
	m.IncrementRateLimitedRequests()
	m.IncrementRateLimitedRequests()
	assert.Equal(t, int64(2), m.GetRateLimitedRequests(), "Should have 2 rate limited requests")

	// Test reset
	m.ResetMetrics()
	assert.Equal(t, int64(0), m.GetSuccessfulRequests(), "Should have no successful requests after reset")
	assert.Equal(t, int64(0), m.GetFailedRequests(), "Should have no failed requests after reset")
	assert.Equal(t, int64(0), m.GetRateLimitedRequests(), "Should have no rate limited requests after reset")
}

func TestHTTPRequestMetricsConcurrently(t *testing.T) {
	m := metrics.NewMetrics()

	// Start goroutines to update metrics
	go func() {
		m.IncrementSuccessfulRequests()
	}()
	go func() {
		m.IncrementFailedRequests()
	}()
	go func() {
		m.IncrementRateLimitedRequests()
	}()

	// Wait for goroutines to complete
	time.Sleep(constants.DefaultTestSleepDuration)

	// Verify metrics
	assert.Equal(t, int64(1), m.GetSuccessfulRequests(), "Should have 1 successful request")
	assert.Equal(t, int64(1), m.GetFailedRequests(), "Should have 1 failed request")
	assert.Equal(t, int64(1), m.GetRateLimitedRequests(), "Should have 1 rate limited request")
}
