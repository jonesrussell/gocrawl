// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/mock"
)

// MockCrawler implements the crawler interface for testing.
type MockCrawler struct {
	mock.Mock
}

// Start implements Interface.
func (m *MockCrawler) Start(ctx context.Context, sourceName string) error {
	args := m.Called(ctx, sourceName)
	return args.Error(0)
}

// Stop implements Interface.
func (m *MockCrawler) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Wait implements Interface.
func (m *MockCrawler) Wait() {
	m.Called()
}

// GetIndexManager implements Interface.
func (m *MockCrawler) GetIndexManager() interfaces.IndexManager {
	args := m.Called()
	if im, ok := args.Get(0).(interfaces.IndexManager); ok {
		return im
	}
	return nil
}

// GetLogger implements Interface.
func (m *MockCrawler) GetLogger() logger.Interface {
	args := m.Called()
	if l, ok := args.Get(0).(logger.Interface); ok {
		return l
	}
	return nil
}

// GetSource implements Interface.
func (m *MockCrawler) GetSource() sources.Interface {
	args := m.Called()
	if s, ok := args.Get(0).(sources.Interface); ok {
		return s
	}
	return nil
}

// GetProcessors implements Interface.
func (m *MockCrawler) GetProcessors() []common.Processor {
	args := m.Called()
	if p, ok := args.Get(0).([]common.Processor); ok {
		return p
	}
	return nil
}

// GetArticleChannel implements Interface.
func (m *MockCrawler) GetArticleChannel() chan *models.Article {
	args := m.Called()
	if c, ok := args.Get(0).(chan *models.Article); ok {
		return c
	}
	return nil
}

// Done implements Interface.
func (m *MockCrawler) Done() <-chan struct{} {
	args := m.Called()
	if c, ok := args.Get(0).(<-chan struct{}); ok {
		return c
	}
	return nil
}

// NewMockCrawler creates a new mock crawler instance.
func NewMockCrawler() *MockCrawler {
	return &MockCrawler{}
}

// Subscribe adds a content handler to receive discovered content.
func (m *MockCrawler) Subscribe(handler events.EventHandler) {
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
