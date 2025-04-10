package testutils

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

// MockCrawler is a mock implementation of the crawler.Interface for testing.
type MockCrawler struct {
	mock.Mock
	Logger           logger.Interface
	indexManager     storagetypes.IndexManager
	sources          sources.Interface
	articleProcessor common.Processor
	contentProcessor common.Processor
	bus              *events.Bus
}

// NewMockCrawler creates a new mock crawler instance.
func NewMockCrawler(
	logger logger.Interface,
	indexManager storagetypes.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
) crawler.Interface {
	return &MockCrawler{
		Logger:           logger,
		indexManager:     indexManager,
		sources:          sources,
		articleProcessor: articleProcessor,
		contentProcessor: contentProcessor,
		bus:              bus,
	}
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

func (m *MockCrawler) GetIndexManager() storagetypes.IndexManager {
	return m.indexManager
}

func (m *MockCrawler) Wait() {
	m.Called()
}

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

// SetTestServerURL sets the test server URL for testing purposes
func (m *MockCrawler) SetTestServerURL(url string) {
	m.Called(url)
}

// Ensure MockCrawler implements crawler.Interface
var _ crawler.Interface = (*MockCrawler)(nil)
