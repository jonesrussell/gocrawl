package crawler_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

func TestNewCrawler(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := storage.NewMockStorage()
	config := &config.Config{
		Crawler: config.CrawlerConfig{
			BaseURL:   "http://example.com",
			MaxDepth:  3,
			RateLimit: time.Second,
			IndexName: "index",
		},
	}

	params := crawler.Params{
		Logger:   mockLogger,
		Storage:  mockStorage,
		Debugger: logger.NewCollyDebugger(mockLogger),
		Config:   config,
	}

	// Set expectations for the Info method
	mockLogger.On("Info", "Crawler initialized", mock.Anything).Return() // Updated to match actual log message

	result, err := crawler.NewCrawler(params)

	require.NoError(t, err)
	require.NotNil(t, result.Crawler)
	require.Equal(t, "http://example.com", result.Crawler.Config.Crawler.BaseURL)
	require.Equal(t, 3, result.Crawler.Config.Crawler.MaxDepth)
	require.Equal(t, time.Second, result.Crawler.Config.Crawler.RateLimit)
}

func TestNewCrawler_MissingLogger(t *testing.T) {
	params := crawler.Params{
		Storage: &storage.MockStorage{},
		Config: &config.Config{
			Crawler: config.CrawlerConfig{
				BaseURL:   "http://example.com",
				MaxDepth:  3,
				RateLimit: time.Second,
				IndexName: "index",
			},
		},
	}

	result, err := crawler.NewCrawler(params)

	require.Error(t, err)
	require.Equal(t, "logger is required", err.Error())
	require.Empty(t, result)
}
