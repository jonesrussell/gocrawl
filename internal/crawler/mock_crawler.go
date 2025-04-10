// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

// MockCrawler is a mock implementation of the crawler interface
type MockCrawler struct {
	mock.Mock
}

// Start mocks the Start method
func (m *MockCrawler) Start(ctx context.Context, sourceName string) error {
	args := m.Called(ctx, sourceName)
	return args.Error(0)
}

// Stop mocks the Stop method
func (m *MockCrawler) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Subscribe mocks the Subscribe method
func (m *MockCrawler) Subscribe(handler events.Handler) {
	m.Called(handler)
}

// SetRateLimit mocks the SetRateLimit method
func (m *MockCrawler) SetRateLimit(duration time.Duration) error {
	args := m.Called(duration)
	return args.Error(0)
}

// SetMaxDepth mocks the SetMaxDepth method
func (m *MockCrawler) SetMaxDepth(depth int) {
	m.Called(depth)
}

// SetCollector mocks the SetCollector method
func (m *MockCrawler) SetCollector(collector *colly.Collector) {
	m.Called(collector)
}

// GetIndexManager returns the index manager.
func (m *MockCrawler) GetIndexManager() storagetypes.IndexManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	if im, ok := args.Get(0).(storagetypes.IndexManager); ok {
		return im
	}
	return nil
}

// Wait mocks the Wait method
func (m *MockCrawler) Wait() {
	m.Called()
}

// GetMetrics returns the crawler metrics.
func (m *MockCrawler) GetMetrics() *common.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	if metrics, ok := args.Get(0).(*common.Metrics); ok {
		return metrics
	}
	return nil
}

// SetTestServerURL mocks the SetTestServerURL method
func (m *MockCrawler) SetTestServerURL(url string) {
	m.Called(url)
}

var _ Interface = (*MockCrawler)(nil)
