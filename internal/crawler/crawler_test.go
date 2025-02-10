package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
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
	mockStorage := &storage.MockStorage{}
	mockLogger := &logger.MockLogger{}
	testConfig := &config.Config{
		IndexName: "test-index",
		BaseURL:   "http://example.com",
		MaxDepth:  1,
	}

	// Set up mock expectations
	mockStorage.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)

	params := crawler.Params{
		BaseURL:   testConfig.BaseURL,
		MaxDepth:  testConfig.MaxDepth,
		RateLimit: 1 * time.Second,
		Debugger:  &logger.CollyDebugger{Logger: mockLogger},
		Logger:    mockLogger,
		Config:    testConfig,
		Storage:   mockStorage,
	}

	crawler, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, crawler.Crawler)
}

// TestCrawler_Start tests the Crawler's Start method
func TestCrawler_Start(t *testing.T) {
	// Setup
	mockLogger := &logger.MockLogger{}
	mockStorage := &storage.MockStorage{}
	mockArticleSvc := &article.MockService{}
	mockShutdowner := &MockShutdowner{}
	mockConfig := &config.Config{
		IndexName: "test_articles",
		BaseURL:   "https://test.com",
		MaxDepth:  1,
	}

	// Create test collector with allowed domains
	c := colly.NewCollector(
		colly.AllowedDomains("www.elliotlaketoday.com", "test.com", "example.com"),
		colly.MaxDepth(1),
	)

	params := crawler.Params{
		BaseURL:   mockConfig.BaseURL,
		MaxDepth:  mockConfig.MaxDepth,
		RateLimit: time.Second,
		Debugger:  &logger.CollyDebugger{Logger: mockLogger},
		Logger:    mockLogger,
		Config:    mockConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)

	crawlerInstance := result.Crawler
	crawlerInstance.SetCollector(c)
	crawlerInstance.SetArticleService(mockArticleSvc)

	// Set up expectations
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockStorage.On("IndexExists", mock.Anything, mockConfig.IndexName).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockShutdowner.On("Shutdown", mock.Anything).Return(nil)

	testArticle := &models.Article{
		ID:    "test-id",
		Title: "Test Article",
		Body:  "Test body",
	}
	mockArticleSvc.On("ExtractArticle", mock.Anything).Return(testArticle)
	mockStorage.On("BulkIndex", mock.Anything, mockConfig.IndexName, mock.Anything).Return(nil)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = crawlerInstance.Start(ctx, mockShutdowner)
	assert.NoError(t, err)

	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockArticleSvc.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}

func TestCrawlerArticleProcessing(t *testing.T) {
	// Setup
	mockLogger := &logger.MockLogger{}
	mockStorage := &storage.MockStorage{}
	mockArticleSvc := &article.MockService{}
	mockShutdowner := &MockShutdowner{}
	mockConfig := &config.Config{
		IndexName: "test_articles",
		BaseURL:   "https://test.com",
		MaxDepth:  1,
	}

	params := crawler.Params{
		BaseURL:   mockConfig.BaseURL,
		MaxDepth:  mockConfig.MaxDepth,
		RateLimit: time.Second,
		Debugger:  &logger.CollyDebugger{Logger: mockLogger},
		Logger:    mockLogger,
		Config:    mockConfig,
		Storage:   mockStorage,
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, html)
	}))
	defer ts.Close()

	// Update BaseURL to test server
	crawlerInstance.BaseURL = ts.URL

	// Set up mock article service
	crawlerInstance.SetArticleService(mockArticleSvc)

	// Set up expectations
	mockStorage.On("IndexExists", mock.Anything, mockConfig.IndexName).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockShutdowner.On("Shutdown", mock.Anything).Return(nil)

	testArticle := &models.Article{
		ID:    "test-id",
		Title: "Test Article",
		Body:  "Test intro\n\nTest body",
	}
	mockArticleSvc.On("ExtractArticle", mock.Anything).Return(testArticle)
	mockStorage.On("BulkIndex", mock.Anything, mockConfig.IndexName, mock.Anything).Return(nil)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = crawlerInstance.Start(ctx, mockShutdowner)
	assert.NoError(t, err)

	// Verify expectations
	mockLogger.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockArticleSvc.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}

func TestCrawler(t *testing.T) {
	mockLogger := logger.NewMockCustomLogger()
	mockStorage := &storage.MockStorage{}
	mockArticleSvc := &article.MockService{}
	mockShutdowner := &MockShutdowner{}

	testConfig := &config.Config{
		IndexName: "test-index",
		BaseURL:   "http://example.com",
		MaxDepth:  1,
	}

	params := crawler.Params{
		BaseURL:   testConfig.BaseURL,
		MaxDepth:  testConfig.MaxDepth,
		RateLimit: 1 * time.Second,
		Logger:    mockLogger,
		Config:    testConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	require.NoError(t, err)
	require.NotNil(t, result)

	crawlerInstance := result.Crawler

	// Set up test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><article>Test content</article></body></html>`)
	}))
	defer ts.Close()

	// Update crawler with test server URL
	crawlerInstance.BaseURL = ts.URL

	// Set up mock article service
	crawlerInstance.SetArticleService(mockArticleSvc)

	// Set up expectations
	mockStorage.On("IndexExists", mock.Anything, testConfig.IndexName).Return(true, nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("BulkIndex", mock.Anything, testConfig.IndexName, mock.Anything).Return(nil)

	testArticle := &models.Article{
		ID:    "test-id",
		Title: "Test Article",
		Body:  "Test content",
	}
	mockArticleSvc.On("ExtractArticle", mock.Anything).Return(testArticle)

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = crawlerInstance.Start(ctx, mockShutdowner)
	assert.NoError(t, err)

	// Verify expectations
	mockArticleSvc.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
