package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
	// Create test dependencies
	log := logger.NewMockCustomLogger()
	mockStorage := storage.NewMockStorage()

	// Create test config
	testConfig := &config.Config{
		Crawler: config.CrawlerConfig{
			IndexName: "test-index",
			BaseURL:   "http://test.com",
			MaxDepth:  3,
			RateLimit: time.Second,
		},
	}

	// Create crawler params
	params := crawler.Params{
		Logger:   log,
		Config:   testConfig,
		Storage:  mockStorage,
		Debugger: nil, // Set this if you have a debugger
	}

	// Test crawler creation
	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, result.Crawler)

	// Verify crawler configuration using getter methods
	require.Equal(t, testConfig.Crawler.BaseURL, result.Crawler.Config.Crawler.BaseURL)
	require.Equal(t, testConfig.Crawler.MaxDepth, result.Crawler.Config.Crawler.MaxDepth)
	require.Equal(t, mockStorage, result.Crawler.Storage)
}

// TestCrawler_Start tests the Crawler's Start method
func TestCrawler_Start(t *testing.T) {
	// Create mock dependencies
	log := logger.NewMockCustomLogger()
	mockStorage := storage.NewMockStorage()

	// Create mock config
	mockConfig := &config.Config{
		Crawler: config.CrawlerConfig{
			IndexName: "test-index",
			BaseURL:   "http://test.com",
			MaxDepth:  3,
			RateLimit: time.Second,
		},
	}

	// Create crawler params
	params := crawler.Params{
		Logger:   log,
		Config:   mockConfig,
		Storage:  mockStorage,
		Debugger: nil, // Set this if you have a debugger
	}

	// Create crawler
	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, result.Crawler)

	// Test crawler start
	ctx := context.Background()
	err = result.Crawler.Start(ctx)
	require.NoError(t, err)

	// Verify storage operations using mock methods
	mockStorage.AssertCalled(t, "TestConnection", ctx)
	mockStorage.AssertCalled(t, "IndexExists", ctx, mockConfig.Crawler.IndexName)

	// Additional tests...
}

func TestCrawlerArticleProcessing(t *testing.T) {
	// Setup
	mockLogger := &logger.MockLogger{}
	mockStorage := &storage.MockStorage{}
	mockArticleSvc := &article.MockService{}

	mockConfig := &config.Config{
		Crawler: config.CrawlerConfig{
			IndexName: "test_articles",
			BaseURL:   "http://example.com",
			MaxDepth:  1,
			RateLimit: time.Second,
		},
	}

	params := crawler.Params{
		Logger:   mockLogger,
		Config:   mockConfig,
		Storage:  mockStorage,
		Debugger: nil, // Set this if you have a debugger
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)

	crawlerInstance := result.Crawler

	// Create test HTML
	html := `
		<html>
			<body>
				<h1 class="details-title">Test Article</h1>
				<div class="details-intro">Test intro</div>
				<div id="details-body">Test body</div>
			</body>
		</html>
	`

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, html)
	}))
	defer ts.Close()

	// Update BaseURL to test server through Config
	crawlerInstance.Config.Crawler.BaseURL = ts.URL

	// Set up mock article service
	crawlerInstance.SetArticleService(mockArticleSvc)

	// Set up expectations
	mockStorage.On("IndexExists", mock.Anything, mockConfig.Crawler.IndexName).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)

	testArticle := &models.Article{
		ID:    "test-id",
		Title: "Test Article",
		Body:  "Test intro\n\nTest body",
	}
	mockArticleSvc.On("ExtractArticle", mock.Anything).Return(testArticle)
	mockStorage.On("BulkIndex", mock.Anything, mockConfig.Crawler.IndexName, mock.Anything).Return(nil)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = crawlerInstance.Start(ctx)
	require.NoError(t, err)

	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockArticleSvc.AssertExpectations(t)
}

func TestCrawler(t *testing.T) {
	mockLogger := logger.NewMockCustomLogger()
	mockStorage := &storage.MockStorage{}
	mockArticleSvc := &article.MockService{}

	testConfig := &config.Config{
		Crawler: config.CrawlerConfig{
			IndexName: "test-index",
			BaseURL:   "http://example.com",
			MaxDepth:  1,
			RateLimit: time.Second,
		},
	}

	params := crawler.Params{
		Logger:   mockLogger,
		Config:   testConfig,
		Storage:  mockStorage,
		Debugger: nil, // Set this if you have a debugger
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, result)

	crawlerInstance := result.Crawler

	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><body><article>Test content</article></body></html>`)
	}))
	defer ts.Close()

	// Update crawler with test server URL through Config
	crawlerInstance.Config.Crawler.BaseURL = ts.URL

	// Set up mock article service
	crawlerInstance.SetArticleService(mockArticleSvc)

	// Set up expectations
	mockStorage.On("IndexExists", mock.Anything, testConfig.Crawler.IndexName).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("BulkIndex", mock.Anything, testConfig.Crawler.IndexName, mock.Anything).Return(nil)

	testArticle := &models.Article{
		ID:    "test-id",
		Title: "Test Article",
		Body:  "Test content",
	}
	mockArticleSvc.On("ExtractArticle", mock.Anything).Return(testArticle)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = crawlerInstance.Start(ctx)
	require.NoError(t, err)

	// Verify expectations
	mockArticleSvc.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
