package crawler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// MockShutdowner is a mock implementation of the fx.Shutdowner interface
type MockShutdowner struct {
	mock.Mock
}

func (m *MockShutdowner) Shutdown(opts ...fx.ShutdownOption) error {
	args := m.Called(opts)
	return args.Error(0)
}

// MockCollector is a mock implementation of colly.Collector
type MockCollector struct {
	mock.Mock
}

func (m *MockCollector) Visit(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

// TestNewCrawler tests the creation of a new Crawler instance
func TestNewCrawler(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	testLogger := logger.NewMockCustomLogger()
	testConfig := &config.Config{IndexName: "test-index"}

	// Set up mock expectations
	mockStorage.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)

	params := crawler.Params{
		BaseURL:   "http://example.com",
		MaxDepth:  1,
		RateLimit: 1 * time.Second,
		Debugger:  &logger.CollyDebugger{},
		Logger:    testLogger,
		Config:    testConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, result.Crawler)
}

// TestCrawler_Start tests the Crawler's Start method
func TestCrawler_Start(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><article><h1>Test Title</h1>Test content</article></body></html>`))
	}))
	defer ts.Close()

	mockStorage := storage.NewMockStorage()
	testLogger := logger.NewMockCustomLogger()
	testConfig := &config.Config{
		IndexName: "test-index",
		CrawlerConfig: config.CrawlerConfig{
			BaseURL:   ts.URL,
			MaxDepth:  1,
			RateLimit: time.Millisecond,
		},
	}

	// Set up mock expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("IndexExists", mock.Anything, testConfig.IndexName).Return(true, nil)
	mockStorage.On("BulkIndexArticles", mock.Anything, mock.Anything).Return(nil)

	params := crawler.Params{
		BaseURL:   ts.URL,
		MaxDepth:  1,
		RateLimit: time.Millisecond,
		Logger:    testLogger,
		Config:    testConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)

	mockShutdowner := new(MockShutdowner)
	mockShutdowner.On("Shutdown", mock.Anything).Return(nil)

	// Create a context with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run the crawler
	err = result.Crawler.Start(ctx, mockShutdowner)
	require.NoError(t, err)

	// Give some time for article processing
	time.Sleep(100 * time.Millisecond)

	// Verify expectations
	mockStorage.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}

func TestCrawlerArticleProcessing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<article>
						<h1>Test Article</h1>
						<p>Test content</p>
						<meta name="author" content="Test Author">
						<meta property="article:tag" content="test">
					</article>
				</body>
			</html>
		`))
	}))
	defer ts.Close()

	mockStorage := storage.NewMockStorage()
	testLogger := logger.NewMockCustomLogger()
	testConfig := &config.Config{
		IndexName: "test-index",
		CrawlerConfig: config.CrawlerConfig{
			BaseURL:   ts.URL,
			MaxDepth:  1,
			RateLimit: time.Millisecond,
		},
	}

	// Set up mock expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("IndexExists", mock.Anything, testConfig.IndexName).Return(true, nil)
	mockStorage.On("BulkIndexArticles", mock.Anything, mock.Anything).Return(nil)

	params := crawler.Params{
		BaseURL:   ts.URL,
		MaxDepth:  1,
		RateLimit: time.Millisecond,
		Logger:    testLogger,
		Config:    testConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)

	mockShutdowner := new(MockShutdowner)
	mockShutdowner.On("Shutdown", mock.Anything).Return(nil)

	// Create a context with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start crawler
	err = result.Crawler.Start(ctx, mockShutdowner)
	require.NoError(t, err)

	// Give some time for article processing
	time.Sleep(100 * time.Millisecond)

	// Verify storage calls
	mockStorage.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}
