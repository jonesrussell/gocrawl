package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/stretchr/testify/mock"
)

// MockCrawler is a mock implementation of crawler.Interface
type MockCrawler struct {
	mock.Mock
}

// Start implements crawler.Interface
func (m *MockCrawler) Start(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
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

// GetIndexManager implements crawler.Interface
func (m *MockCrawler) GetIndexManager() api.IndexManager {
	args := m.Called()
	if result := args.Get(0); result != nil {
		if im, ok := result.(api.IndexManager); ok {
			return im
		}
	}
	return nil
}

// Wait implements crawler.Interface
func (m *MockCrawler) Wait() {
	m.Called()
}

// GetMetrics implements crawler.Interface
func (m *MockCrawler) GetMetrics() *collector.Metrics {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	if metrics, ok := args.Get(0).(*collector.Metrics); ok {
		return metrics
	}
	return nil
}

// Ensure MockCrawler implements crawler.Interface
var _ Interface = (*MockCrawler)(nil)
