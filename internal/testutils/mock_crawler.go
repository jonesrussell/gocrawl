package testutils

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
)

// MockCrawler implements crawler.Interface for testing
type MockCrawler struct {
	crawler.Interface
}

// NewMockCrawler creates a new mock crawler instance
func NewMockCrawler() *MockCrawler {
	return &MockCrawler{}
}

func (m *MockCrawler) Start(_ context.Context, _ string) error {
	return nil
}

func (m *MockCrawler) Stop(_ context.Context) error {
	return nil
}

func (m *MockCrawler) Subscribe(_ events.Handler) {
}

func (m *MockCrawler) SetRateLimit(_ time.Duration) error {
	return nil
}

func (m *MockCrawler) SetMaxDepth(_ int) {
}

func (m *MockCrawler) SetCollector(_ *colly.Collector) {
}

func (m *MockCrawler) GetIndexManager() api.IndexManager {
	return NewMockIndexManager()
}

func (m *MockCrawler) Wait() {
}

func (m *MockCrawler) GetMetrics() *collector.Metrics {
	return &collector.Metrics{}
}
