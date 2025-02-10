package crawler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) IndexDocument(ctx context.Context, indexName, docID string, document interface{}) error {
	args := m.Called(ctx, indexName, docID, document)
	return args.Error(0)
}

func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockShutdowner is a mock implementation of the fx.Shutdowner interface
type MockShutdowner struct {
	mock.Mock
}

func (m *MockShutdowner) Shutdown(opts ...fx.ShutdownOption) error {
	args := m.Called(opts)
	return args.Error(0)
}

// TestNewCrawler tests the creation of a new Crawler instance
func TestNewCrawler(t *testing.T) {
	mockStorage := new(MockStorage)

	// Use NewMockCustomLogger to create a mock logger
	testLogger := logger.NewMockCustomLogger()

	testConfig := &config.Config{IndexName: "test-index"}

	params := crawler.Params{
		BaseURL:   "http://example.com",
		MaxDepth:  1,
		RateLimit: 1 * time.Second,
		Debugger:  &logger.CollyDebugger{},
		Logger:    testLogger, // Use the assigned logger variable
		Config:    testConfig,
		Storage:   mockStorage,
	}

	result, err := crawler.NewCrawler(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Crawler == nil {
		t.Fatalf("Expected crawler instance, got nil")
	}
}

// TestCrawler_Start tests the Crawler's Start method
func TestCrawler_Start(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Test content</body></html>`))
	}))
	defer ts.Close()

	mockStorage := new(MockStorage)
	testLogger := logger.NewMockCustomLogger()
	testConfig := &config.Config{IndexName: "test-index"}

	// Create a new collector without domain restrictions
	collector := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
		colly.Async(true),
		colly.AllowedDomains(), // Empty to allow all domains
	)

	c := &crawler.Crawler{
		BaseURL:   ts.URL,
		Storage:   mockStorage,
		MaxDepth:  1,
		RateLimit: 1 * time.Millisecond,
		Collector: collector,
		Logger:    testLogger,
		IndexName: testConfig.IndexName,
	}

	// Set up mock expectations
	mockStorage.On("TestConnection", mock.Anything).Return(nil)
	mockStorage.On("IndexDocument", mock.Anything, testConfig.IndexName, mock.Anything, mock.Anything).Return(nil).Maybe()

	mockShutdowner := new(MockShutdowner)
	mockShutdowner.On("Shutdown", mock.Anything).Return(nil)

	app := fxtest.New(t)
	defer app.RequireStart().RequireStop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run the crawler
	if err := c.Start(ctx, mockShutdowner); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify expectations
	mockStorage.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}
