package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

// MockShutdowner is a simple mock for the fx.Shutdowner
type MockShutdowner struct{}

func (m *MockShutdowner) Shutdown(...fx.ShutdownOption) error {
	return nil
}

// TestNewCrawler tests the NewCrawler function
func TestNewCrawler(t *testing.T) {
	// Create logger parameters
	loggerParams := logger.Params{
		Level: zapcore.InfoLevel, // Set the desired log level
	}

	mockLogger, err := logger.NewCustomLogger(loggerParams) // Pass the logger parameters
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	cfg := &config.Config{
		AppName:    "TestApp",
		AppEnv:     "development",
		LogLevel:   "DEBUG",
		ElasticURL: "http://localhost:9200", // This can be ignored in mock
		IndexName:  "test_index",
	}

	// Use MockStorage instead of real storage
	mockStorage := storage.NewMockStorage()

	params := crawler.Params{
		BaseURL:   "http://example.com",
		MaxDepth:  2,
		RateLimit: 1 * time.Second,
		Logger:    mockLogger,
		Config:    cfg,
		Storage:   mockStorage, // Pass the mock storage
	}

	result, err := crawler.NewCrawler(params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Crawler == nil {
		t.Fatal("expected a non-nil Crawler")
	}
	if result.Crawler.BaseURL != params.BaseURL {
		t.Errorf("expected BaseURL %s, got %s", params.BaseURL, result.Crawler.BaseURL)
	}
}

// TestCrawlerStart tests the Start method of the Crawler
func TestCrawlerStart(t *testing.T) {
	mockLogger, err := logger.NewCustomLogger(logger.Params{
		Level: zapcore.InfoLevel, // Set the desired log level
	})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	cfg := &config.Config{
		AppName:    "TestApp",
		AppEnv:     "development",
		LogLevel:   "DEBUG",
		ElasticURL: "http://localhost:9200",
		IndexName:  "test_index",
	}

	params := crawler.Params{
		BaseURL:   "http://example.com",
		MaxDepth:  2,
		RateLimit: 1 * time.Second,
		Logger:    mockLogger,
		Config:    cfg,
		Storage:   storage.NewMockStorage(), // Use MockStorage
	}

	crawlerResult, err := crawler.NewCrawler(params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Mock the shutdowner
	shutdowner := &MockShutdowner{}

	// Call the Start method
	err = crawlerResult.Crawler.Start(context.Background(), shutdowner)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
