package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
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
	return args.Get(0).(api.IndexManager)
}

// Wait implements crawler.Interface
func (m *MockCrawler) Wait() {
	m.Called()
}
