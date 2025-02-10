package crawler_test

import (
	"context"
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
	mockStorage := new(MockStorage)

	// Use NewMockCustomLogger to create a mock logger
	testLogger := logger.NewMockCustomLogger()

	testConfig := &config.Config{IndexName: "test-index"}

	c := &crawler.Crawler{
		BaseURL:   "http://example.com",
		Storage:   mockStorage,
		MaxDepth:  1,
		RateLimit: 1 * time.Second,
		Collector: colly.NewCollector(),
		Logger:    testLogger,
		IndexName: testConfig.IndexName,
	}

	mockStorage.On("IndexDocument", mock.Anything, testConfig.IndexName, mock.Anything, mock.Anything).Return(nil)
	mockStorage.On("TestConnection", mock.Anything).Return(nil)

	// Create a mock shutdowner
	mockShutdowner := new(MockShutdowner)
	mockShutdowner.On("Shutdown").Return(nil)

	app := fxtest.New(t)
	defer app.RequireStart().RequireStop()

	ctx := context.Background()
	err := c.Start(ctx, mockShutdowner)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockStorage.AssertExpectations(t)
	mockShutdowner.AssertExpectations(t)
}
