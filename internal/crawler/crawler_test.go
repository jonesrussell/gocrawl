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
	"go.uber.org/fx/fxtest"
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
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("BulkIndexArticles", mock.Anything, mock.Anything).Return(nil)

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

	mockStorage.AssertExpectations(t)
}

// TestCrawler_Start tests the Crawler's Start method
func TestCrawler_Start(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Test content</body></html>`))
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
	mockStorage.On("IndexDocument", mock.Anything, testConfig.IndexName, mock.Anything, mock.Anything).Return(nil)
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

	app := fxtest.New(t)
	defer app.RequireStart().RequireStop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run the crawler
	err = result.Crawler.Start(ctx, mockShutdowner)
	require.NoError(t, err)

	// Verify expectations
	mockStorage.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}

func TestCrawlerArticleProcessing(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	testLogger := logger.NewMockCustomLogger()
	testConfig := &config.Config{
		IndexName: "test-index",
		CrawlerConfig: config.CrawlerConfig{
			BaseURL:   "https://example.com",
			MaxDepth:  1,
			RateLimit: time.Millisecond,
		},
	}

	// Set up mock expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("BulkIndexArticles", mock.Anything, mock.Anything).Return(nil)

	params := crawler.Params{
		BaseURL:   "https://example.com",
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start crawler
	err = result.Crawler.Start(ctx, mockShutdowner)
	require.NoError(t, err)

	// Verify storage calls
	mockStorage.AssertExpectations(t)
}
