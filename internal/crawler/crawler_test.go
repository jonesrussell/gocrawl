package crawler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
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

// CanProcess implements ContentProcessor.CanProcess
func (m *MockProcessor) CanProcess(content interface{}) bool {
	_, ok := content.(*colly.HTMLElement)
	return ok
}

// ContentType implements ContentProcessor.ContentType
func (m *MockProcessor) ContentType() common.ContentType {
	return common.ContentTypePage
}

// Ensure MockProcessor implements common.Processor
var _ common.Processor = (*MockProcessor)(nil)

// MockLogger is a mock implementation of common.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Printf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
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

// MockSources is a mock implementation of sources.Interface
type MockSources struct {
	mock.Mock
}

func (m *MockSources) FindByName(name string) (*sources.Config, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sources.Config), args.Error(1)
}

func (m *MockSources) AddSource(ctx context.Context, source *sources.Config) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) ListSources(ctx context.Context) ([]*sources.Config, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*sources.Config), args.Error(1)
}

func (m *MockSources) UpdateSource(ctx context.Context, source *sources.Config) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) DeleteSource(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockSources) ValidateSource(source *sources.Config) error {
	args := m.Called(source)
	return args.Error(0)
}

func (m *MockSources) GetMetrics() sources.Metrics {
	args := m.Called()
	return args.Get(0).(sources.Metrics)
}

func (m *MockSources) GetSources() ([]sources.Config, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sources.Config), args.Error(1)
}

// NewCrawler creates a new crawler instance with the given dependencies
func NewCrawler(
	logger common.Logger,
	bus *events.Bus,
	indexManager api.IndexManager,
	sources sources.Interface,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
) crawler.Interface {
	return crawler.NewCrawler(
		logger,
		indexManager,
		sources,
		articleProcessor,
		contentProcessor,
		bus,
	)
}

// TestCrawlerStartup tests crawler startup functionality.
func TestCrawlerStartup(t *testing.T) {
	// Create test logger
	testLogger, initErr := logger.NewCustomLogger(nil, logger.Params{
		Debug:  true,
		Level:  "info",
		AppEnv: "development",
	})
	require.NoError(t, initErr)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Test Page</h1></body></html>"))
	}))
	defer server.Close()

	// Create mock sources with the test server URL
	mockSources := &sources.Sources{}
	mockSources.SetSources([]sources.Config{
		{
			Name:      "test",
			URL:       "http://test.example.com", // Use the test domain from config
			RateLimit: time.Millisecond * 100,    // Use a shorter rate limit for testing
			MaxDepth:  1,
		},
	})

	// Initialize collector with test configuration
	collector := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)

	// Configure collector for testing
	collector.AllowedDomains = []string{"test.example.com"} // Use the test domain from config
	collector.DisallowedDomains = nil                       // Don't disallow any domains
	collector.MaxDepth = 1
	collector.DetectCharset = true
	collector.CheckHead = true
	collector.AllowURLRevisit = true
	collector.Async = true

	// Log collector configuration
	testLogger.Info("Configured collector for testing",
		"allowed_domains", collector.AllowedDomains,
		"disallowed_domains", collector.DisallowedDomains,
		"max_depth", collector.MaxDepth,
		"allow_url_revisit", collector.AllowURLRevisit,
		"async", collector.Async)

	// Set up callbacks for testing
	collector.OnRequest(func(r *colly.Request) {
		testLogger.Info("Visiting", "url", r.URL.String())
	})

	collector.OnResponse(func(r *colly.Response) {
		testLogger.Info("Visited", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	collector.OnError(func(r *colly.Response, err error) {
		testLogger.Error("Error while crawling",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
			"error", err)
	})

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources with test data
			func() sources.Interface { return mockSources },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus with correct name
			fx.Annotate(
				events.NewBus,
				fx.ResultTags(`name:"eventBus"`),
			),
		),
		fx.Invoke(func(c crawler.Interface) {
			c.SetCollector(collector)
			c.SetTestServerURL(server.URL) // Set the test server URL

			// Test startup with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			startErr := c.Start(ctx, "test")
			require.NoError(t, startErr)

			// Wait for crawler to finish
			c.Wait()
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestCrawlerShutdown tests crawler shutdown functionality.
func TestCrawlerShutdown(t *testing.T) {
	// Create test logger
	testLogger, initErr := logger.NewCustomLogger(nil, logger.Params{
		Debug:  true,
		Level:  "info",
		AppEnv: "development",
	})
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources with test data
			func() sources.Interface { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus with correct name
			fx.Annotate(
				events.NewBus,
				fx.ResultTags(`name:"eventBus"`),
			),
		),
		fx.Invoke(func(c crawler.Interface) {
			// Initialize collector
			collector := colly.NewCollector()
			c.SetCollector(collector)

			// Test shutdown with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			stopErr := c.Stop(ctx)
			require.NoError(t, stopErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestSourceValidation tests source validation functionality.
func TestSourceValidation(t *testing.T) {
	// Create test logger
	testLogger, initErr := logger.NewCustomLogger(nil, logger.Params{
		Debug:  true,
		Level:  "info",
		AppEnv: "development",
	})
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources with test data
			func() sources.Interface { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus with correct name
			fx.Annotate(
				events.NewBus,
				fx.ResultTags(`name:"eventBus"`),
			),
		),
		fx.Invoke(func(c crawler.Interface) {
			// Initialize collector
			collector := colly.NewCollector()
			c.SetCollector(collector)

			// Test source validation with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			startErr := c.Start(ctx, "nonexistent")
			require.Error(t, startErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// TestErrorHandling tests error handling functionality.
func TestErrorHandling(t *testing.T) {
	// Create test logger
	testLogger, initErr := logger.NewCustomLogger(nil, logger.Params{
		Debug:  true,
		Level:  "info",
		AppEnv: "development",
	})
	require.NoError(t, initErr)

	// Create test app with all required dependencies
	app := fx.New(
		crawler.Module,
		fx.Provide(
			// Provide logger
			func() common.Logger { return testLogger },
			// Provide debugger
			func() debug.Debugger {
				return &debug.LogDebugger{
					Output: crawler.NewDebugLogger(testLogger),
				}
			},
			// Provide index manager
			func() api.IndexManager { return &mockIndexManager{} },
			// Provide sources with test data
			func() sources.Interface { return &sources.Sources{} },
			// Provide article processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupArticleProcessor"`),
			),
			// Provide content processor with correct name
			fx.Annotate(
				func() common.Processor { return &MockProcessor{} },
				fx.ResultTags(`name:"startupContentProcessor"`),
			),
			// Provide event bus with correct name
			fx.Annotate(
				events.NewBus,
				fx.ResultTags(`name:"eventBus"`),
			),
		),
		fx.Invoke(func(c crawler.Interface) {
			// Initialize collector
			collector := colly.NewCollector()
			c.SetCollector(collector)

			// Test error handling with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			startErr := c.Start(ctx, "nonexistent")
			require.Error(t, startErr)
		}),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
}

// writerWrapper implements io.Writer for the logger
type writerWrapper struct {
	logger common.Logger
}

// Write implements io.Writer interface
func (w *writerWrapper) Write(p []byte) (int, error) {
	w.logger.Debug(string(p))
	return len(p), nil
}

// NewDebugLogger creates a debug logger for testing.
func NewDebugLogger(logger common.Logger) io.Writer {
	return &writerWrapper{logger: logger}
}

func TestCrawler_ProcessHTML(t *testing.T) {
	// Create a test context
	ctx := context.Background()

	// Create a development logger with nice formatting
	devLogger, err := logger.NewCustomLogger(nil, logger.Params{
		Debug:  true,
		Level:  "debug",
		AppEnv: "development",
	})
	require.NoError(t, err)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
				<body>
					<article>
						<h1>Test Article</h1>
						<div class="content">
							<p>Test content</p>
						</div>
					</article>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	// Create test dependencies
	mockIndexManager := &MockIndexManager{}
	mockSources := &MockSources{}
	mockArticleProcessor := &MockProcessor{}
	mockContentProcessor := &MockProcessor{}
	bus := events.NewBus()

	// Create crawler with development logger
	c := NewCrawler(
		devLogger,
		bus,
		mockIndexManager,
		mockSources,
		mockArticleProcessor,
		mockContentProcessor,
	)

	// Create a test collector without domain restrictions
	collector := colly.NewCollector(
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)
	c.SetCollector(collector)

	// Set up the crawler's HTML handlers
	collector.OnHTML("article", func(e *colly.HTMLElement) {
		// Set up expectations for the actual element being processed
		mockArticleProcessor.On("ProcessHTML", ctx, e).Return(nil)
		err := mockArticleProcessor.ProcessHTML(ctx, e)
		require.NoError(t, err)
	})

	collector.OnHTML("div.content", func(e *colly.HTMLElement) {
		// Set up expectations for the actual element being processed
		mockContentProcessor.On("ProcessHTML", ctx, e).Return(nil)
		err := mockContentProcessor.ProcessHTML(ctx, e)
		require.NoError(t, err)
	})

	// Visit the test server URL
	err = collector.Visit(server.URL)
	require.NoError(t, err)

	// Wait for the crawler to finish
	c.Wait()

	// Verify expectations
	mockArticleProcessor.AssertExpectations(t)
	mockContentProcessor.AssertExpectations(t)
}
