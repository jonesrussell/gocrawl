package testutils

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
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
	indexManager   interfaces.IndexManager
	source         sources.Interface
	processors     []common.Processor
	articleChannel chan *models.Article
	logger         logger.Interface
}

// NewMockCrawler creates a new mock crawler instance.
func NewMockCrawler(
	indexManager interfaces.IndexManager,
	source sources.Interface,
	processors []common.Processor,
	articleChannel chan *models.Article,
	logger logger.Interface,
) *MockCrawler {
	return &MockCrawler{
		indexManager:   indexManager,
		source:         source,
		processors:     processors,
		articleChannel: articleChannel,
		logger:         logger,
	}
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
	return m.indexManager
}

// GetLogger implements Interface.
func (m *MockCrawler) GetLogger() logger.Interface {
	return m.logger
}

// GetSource implements Interface.
func (m *MockCrawler) GetSource() sources.Interface {
	return m.source
}

// GetProcessors implements Interface.
func (m *MockCrawler) GetProcessors() []common.Processor {
	return m.processors
}

// GetArticleChannel implements Interface.
func (m *MockCrawler) GetArticleChannel() chan *models.Article {
	return m.articleChannel
}

// Done implements Interface.
func (m *MockCrawler) Done() <-chan struct{} {
	args := m.Called()
	if c, ok := args.Get(0).(<-chan struct{}); ok {
		return c
	}
	return nil
}

func (m *MockCrawler) Subscribe(handler events.EventHandler) {
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
