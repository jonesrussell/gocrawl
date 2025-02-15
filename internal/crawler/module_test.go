package crawler_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

	assert.NoError(t, err)
	assert.NotNil(t, result.Crawler)
	assert.Equal(t, "http://example.com", result.Crawler.Config.Crawler.BaseURL)
	assert.Equal(t, 3, result.Crawler.Config.Crawler.MaxDepth)
	assert.Equal(t, time.Second, result.Crawler.Config.Crawler.RateLimit)
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

	assert.Error(t, err)
	assert.Equal(t, "logger is required", err.Error())
	assert.Empty(t, result)
}
