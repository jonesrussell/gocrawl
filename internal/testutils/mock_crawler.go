package testutils

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	"github.com/stretchr/testify/mock"
)

// MockCrawler implements crawler.Interface for testing
type MockCrawler struct {
	mock.Mock
}

// NewMockCrawler creates a new mock crawler instance
func NewMockCrawler() *MockCrawler {
	return &MockCrawler{}
}

func (m *MockCrawler) Start(ctx context.Context, source string) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockCrawler) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCrawler) Subscribe(handler events.Handler) {
	m.Called(handler)
}

func (m *MockCrawler) SetRateLimit(duration time.Duration) error {
	args := m.Called(duration)
	return args.Error(0)
}

func (m *MockCrawler) SetMaxDepth(depth int) {
	m.Called(depth)
}

func (m *MockCrawler) SetCollector(collector *colly.Collector) {
	m.Called(collector)
}

func (m *MockCrawler) GetIndexManager() api.IndexManager {
	args := m.Called()
	if result := args.Get(0); result != nil {
		if im, ok := result.(api.IndexManager); ok {
			return im
		}
	}
	return nil
}

func (m *MockCrawler) Wait() {
	m.Called()
}

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
var _ crawler.Interface = (*MockCrawler)(nil)
