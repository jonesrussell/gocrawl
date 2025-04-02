// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/stretchr/testify/mock"
)

// MockCrawler is a mock implementation of crawler.Interface
type MockCrawler struct {
	mock.Mock
}

// Start implements crawler.Interface
func (m *MockCrawler) Start(ctx context.Context, sourceName string) error {
	args := m.Called(ctx, sourceName)
	return args.Error(0)
}

// Stop implements crawler.Interface
func (m *MockCrawler) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Subscribe implements crawler.Interface
func (m *MockCrawler) Subscribe(handler events.Handler) {
	m.Called(handler)
}

// SetRateLimit implements crawler.Interface
func (m *MockCrawler) SetRateLimit(duration time.Duration) error {
	args := m.Called(duration)
	return args.Error(0)
}

// SetMaxDepth implements crawler.Interface
func (m *MockCrawler) SetMaxDepth(depth int) {
	m.Called(depth)
}

// SetCollector implements crawler.Interface
func (m *MockCrawler) SetCollector(collector *colly.Collector) {
	m.Called(collector)
}

// GetIndexManager returns the index manager interface.
func (m *MockCrawler) GetIndexManager() api.IndexManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	indexManager, ok := args.Get(0).(api.IndexManager)
	if !ok {
		return nil
	}
	return indexManager
}

// Wait implements crawler.Interface
func (m *MockCrawler) Wait() {
	m.Called()
}

// GetMetrics returns the current crawler metrics.
func (m *MockCrawler) GetMetrics() *common.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	metrics, ok := args.Get(0).(*common.Metrics)
	if !ok {
		return nil
	}
	return metrics
}

// Ensure MockCrawler implements crawler.Interface
var _ Interface = (*MockCrawler)(nil)
