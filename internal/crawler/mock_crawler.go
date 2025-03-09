package crawler

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
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
func (m *MockCrawler) Stop() {
	m.Called()
}

// SetCollector implements crawler.Interface
func (m *MockCrawler) SetCollector(collector *colly.Collector) {
	m.Called(collector)
}

// SetService implements crawler.Interface
func (m *MockCrawler) SetService(service article.Interface) {
	m.Called(service)
}

// GetBaseURL implements crawler.Interface
func (m *MockCrawler) GetBaseURL() string {
	args := m.Called()
	return args.String(0)
}

// Wait implements crawler.Interface
func (m *MockCrawler) Wait() {
	m.Called()
}

// NewMockCrawler creates a new instance of MockCrawler
func NewMockCrawler() *MockCrawler {
	return &MockCrawler{}
}
