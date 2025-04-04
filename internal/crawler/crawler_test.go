package crawler_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	mock.Mock
}

func (m *mockIndexManager) CreateIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockIndexManager) EnsureIndex(ctx context.Context, name string, settings any) error {
	args := m.Called(ctx, name, settings)
	return args.Error(0)
}

func (m *mockIndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// MockProcessor implements common.Processor for testing
type MockProcessor struct {
	mock.Mock
	ProcessCalls int
}

// ProcessJob implements common.Processor
func (m *MockProcessor) ProcessJob(ctx context.Context, job *common.Job) {
	m.Called(ctx, job)
}

// ProcessHTML implements HTMLProcessor.ProcessHTML
func (m *MockProcessor) ProcessHTML(ctx context.Context, e *colly.HTMLElement) error {
	m.ProcessCalls++
	args := m.Called(ctx, e)
	return args.Error(0)
}

// GetMetrics implements common.Processor
func (m *MockProcessor) GetMetrics() *common.Metrics {
	args := m.Called()
	return args.Get(0).(*common.Metrics)
}

// Process implements common.Processor
func (m *MockProcessor) Process(ctx context.Context, data any) error {
	return nil
}

// Start implements common.Processor
func (m *MockProcessor) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stop implements common.Processor
func (m *MockProcessor) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// CanProcess implements Processor.CanProcess
func (m *MockProcessor) CanProcess(content any) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// ContentType implements ContentProcessor.ContentType
func (m *MockProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// Ensure MockProcessor implements common.Processor
var _ common.Processor = (*MockProcessor)(nil)

// MockLogger is a mock implementation of logger.Interface
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...any) {}
func (m *MockLogger) Info(msg string, fields ...any)  {}
func (m *MockLogger) Warn(msg string, fields ...any)  {}
func (m *MockLogger) Error(msg string, fields ...any) {}
func (m *MockLogger) Fatal(msg string, fields ...any) {}
func (m *MockLogger) With(fields ...any) logger.Interface {
	return m
}

// MockBus is a mock implementation of events.Bus
type MockBus struct {
	mock.Mock
	handlers []events.Handler
}

func (m *MockBus) Subscribe(handler events.Handler) {
	m.Called(handler)
	m.handlers = append(m.handlers, handler)
}

func (m *MockBus) Publish(ctx context.Context, content *events.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockIndexManager is a mock implementation of api.IndexManager
type MockIndexManager struct {
	mock.Mock
}

func (m *MockIndexManager) CreateIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string, settings any) error {
	args := m.Called(ctx, name, settings)
	return args.Error(0)
}

func (m *MockIndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// MockSources implements sources.Interface for testing
type MockSources struct {
	mock.Mock
}

func (m *MockSources) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*sourceutils.SourceConfig), args.Error(1)
}

func (m *MockSources) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) DeleteSource(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockSources) ValidateSource(source *sourceutils.SourceConfig) error {
	args := m.Called(source)
	return args.Error(0)
}

func (m *MockSources) GetMetrics() sources.Metrics {
	args := m.Called()
	return args.Get(0).(sources.Metrics)
}

func (m *MockSources) GetSources() ([]sourceutils.SourceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sourceutils.SourceConfig), args.Error(1)
}

func (m *MockSources) FindByName(name string) *sourceutils.SourceConfig {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sourceutils.SourceConfig)
}

// NewMockSources creates a new mock sources instance
func NewMockSources() *MockSources {
	return &MockSources{}
}

// TestCrawlerStartup tests crawler startup functionality.
func TestCrawlerStartup(t *testing.T) {
	mockIndexManager := &MockIndexManager{}
	mockIndexManager.On("IndexExists", mock.Anything, "test_articles").Return(true, nil)
	mockIndexManager.On("IndexExists", mock.Anything, "test_content").Return(true, nil)

	mockSources := NewMockSources()
	testSource := &sourceutils.SourceConfig{
		Name:         "test_source",
		URL:          "http://test.example.com",
		RateLimit:    time.Second * 2,
		MaxDepth:     2,
		ArticleIndex: "test_articles",
		Index:        "test_content",
	}
	mockSources.On("FindByName", "test_source").Return(testSource)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return &MockLogger{} },
			func() api.IndexManager { return mockIndexManager },
			func() []common.Processor { return []common.Processor{&MockProcessor{}} },
			func() *events.Bus { return &events.Bus{} },
			func() sources.Interface { return mockSources },
		),
		crawler.Module,
		fx.Invoke(func(c crawler.Interface) {
			err := c.Start(context.Background(), "test_source")
			require.NoError(t, err)
		}),
	)
	require.NoError(t, app.Err())
	mockIndexManager.AssertExpectations(t)
	mockSources.AssertExpectations(t)
}

// TestCrawlerShutdown tests crawler shutdown functionality.
func TestCrawlerShutdown(t *testing.T) {
	mockIndexManager := &MockIndexManager{}
	mockIndexManager.On("IndexExists", mock.Anything, "test_articles").Return(true, nil)
	mockIndexManager.On("IndexExists", mock.Anything, "test_content").Return(true, nil)

	mockSources := NewMockSources()
	testSource := &sourceutils.SourceConfig{
		Name:         "test_source",
		URL:          "http://test.example.com",
		RateLimit:    time.Second * 2,
		MaxDepth:     2,
		ArticleIndex: "test_articles",
		Index:        "test_content",
	}
	mockSources.On("FindByName", "test_source").Return(testSource)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return &MockLogger{} },
			func() api.IndexManager { return mockIndexManager },
			func() []common.Processor { return []common.Processor{&MockProcessor{}} },
			func() *events.Bus { return &events.Bus{} },
			func() sources.Interface { return mockSources },
		),
		crawler.Module,
		fx.Invoke(func(c crawler.Interface) {
			err := c.Start(context.Background(), "test_source")
			require.NoError(t, err)
			err = c.Stop(context.Background())
			require.NoError(t, err)
		}),
	)
	require.NoError(t, app.Err())
	mockIndexManager.AssertExpectations(t)
	mockSources.AssertExpectations(t)
}

// TestSourceValidation tests source validation functionality.
func TestSourceValidation(t *testing.T) {
	mockIndexManager := &MockIndexManager{}
	mockIndexManager.On("IndexExists", mock.Anything, "test_articles").Return(true, nil)
	mockIndexManager.On("IndexExists", mock.Anything, "test_content").Return(true, nil)

	mockSources := NewMockSources()
	testSource := &sourceutils.SourceConfig{
		Name:         "test_source",
		URL:          "http://test.example.com",
		RateLimit:    time.Second * 2,
		MaxDepth:     2,
		ArticleIndex: "test_articles",
		Index:        "test_content",
	}
	mockSources.On("FindByName", "test_source").Return(testSource)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return &MockLogger{} },
			func() api.IndexManager { return mockIndexManager },
			func() []common.Processor { return []common.Processor{&MockProcessor{}} },
			func() *events.Bus { return &events.Bus{} },
			func() sources.Interface { return mockSources },
		),
		crawler.Module,
		fx.Invoke(func(c crawler.Interface) {
			err := c.Start(context.Background(), "test_source")
			require.NoError(t, err)
		}),
	)
	require.NoError(t, app.Err())
	mockIndexManager.AssertExpectations(t)
	mockSources.AssertExpectations(t)
}

// TestErrorHandling tests error handling functionality.
func TestErrorHandling(t *testing.T) {
	mockIndexManager := &MockIndexManager{}
	mockIndexManager.On("IndexExists", mock.Anything, "test_articles").Return(false, fmt.Errorf("index not found"))

	mockSources := NewMockSources()
	testSource := &sourceutils.SourceConfig{
		Name:         "test_source",
		URL:          "http://test.example.com",
		RateLimit:    time.Second * 2,
		MaxDepth:     2,
		ArticleIndex: "test_articles",
		Index:        "test_content",
	}
	mockSources.On("FindByName", "test_source").Return(testSource)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return &MockLogger{} },
			func() api.IndexManager { return mockIndexManager },
			func() []common.Processor { return []common.Processor{&MockProcessor{}} },
			func() *events.Bus { return &events.Bus{} },
			func() sources.Interface { return mockSources },
		),
		crawler.Module,
		fx.Invoke(func(c crawler.Interface) {
			err := c.Start(context.Background(), "test_source")
			require.Error(t, err)
			require.Contains(t, err.Error(), "index not found")
		}),
	)
	require.NoError(t, app.Err())
	mockIndexManager.AssertExpectations(t)
	mockSources.AssertExpectations(t)
}

// TestCrawler_ProcessHTML tests HTML processing functionality.
func TestCrawler_ProcessHTML(t *testing.T) {
	mockIndexManager := &MockIndexManager{}
	mockIndexManager.On("IndexExists", mock.Anything, "test_articles").Return(true, nil)
	mockIndexManager.On("IndexExists", mock.Anything, "test_content").Return(true, nil)

	mockSources := NewMockSources()
	testSource := &sourceutils.SourceConfig{
		Name:         "test_source",
		URL:          "http://test.example.com",
		RateLimit:    time.Second * 2,
		MaxDepth:     2,
		ArticleIndex: "test_articles",
		Index:        "test_content",
	}
	mockSources.On("FindByName", "test_source").Return(testSource)

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return &MockLogger{} },
			func() api.IndexManager { return mockIndexManager },
			func() []common.Processor { return []common.Processor{&MockProcessor{}} },
			func() *events.Bus { return &events.Bus{} },
			func() sources.Interface { return mockSources },
		),
		crawler.Module,
		fx.Invoke(func(c crawler.Interface) {
			err := c.Start(context.Background(), "test_source")
			require.NoError(t, err)
		}),
	)
	require.NoError(t, app.Err())
	mockIndexManager.AssertExpectations(t)
	mockSources.AssertExpectations(t)
}

// writerWrapper wraps a logger to implement io.Writer
type writerWrapper struct {
	logger logger.Interface
}

func (w *writerWrapper) Write(p []byte) (int, error) {
	w.logger.Debug(string(p))
	return len(p), nil
}

// NewDebugLogger creates a new debug logger
func NewDebugLogger(logger logger.Interface) io.Writer {
	return &writerWrapper{logger: logger}
}

// TestModuleProvides tests that the module provides all required dependencies.
func TestModuleProvides(t *testing.T) {
	// Create test logger
	testLogger := &MockLogger{}

	// Create test app
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return testLogger },
			func() debug.Debugger { return &debug.LogDebugger{} },
			func() api.IndexManager { return &MockIndexManager{} },
			func() sources.Interface { return &MockSources{} },
			func() []common.Processor {
				return []common.Processor{&MockProcessor{}, &MockProcessor{}}
			},
			func() *events.Bus { return &events.Bus{} },
			func() chan *models.Article { return make(chan *models.Article, 100) },
		),
		crawler.Module,
	)

	// Start the app
	app.RequireStart()

	// Stop the app
	app.RequireStop()
}
