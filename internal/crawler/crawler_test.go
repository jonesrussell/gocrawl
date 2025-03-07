package crawler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

func TestCrawler_Stop(_ *testing.T) {
	mockLogger := logger.NewMockLogger()
	c := &crawler.Crawler{
		Logger: mockLogger,
	}

	// There is no specific behavior to test for Stop method since it does nothing
	c.Stop()
	// Here we could test if any cleanup code is executed
}

func TestCrawler_SetCollector(t *testing.T) {
	c := &crawler.Crawler{}
	collector := &colly.Collector{}
	c.SetCollector(collector)
	assert.Equal(t, collector, c.Collector)
}

func TestCrawler_SetService(t *testing.T) {
	c := &crawler.Crawler{}
	service := article.NewMockService()
	c.SetService(service)
	assert.Equal(t, service, c.ArticleService)
}

func TestCrawler_GetBaseURL(t *testing.T) {
	configInstance := &config.Config{
		Crawler: config.CrawlerConfig{
			BaseURL: "http://example.com",
		},
	}
	c := &crawler.Crawler{Config: configInstance}
	assert.Equal(t, "http://example.com", c.GetBaseURL())
}

func TestCrawler_Start(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		setupMock func(*storage.MockStorage, *logger.MockLogger, *storage.MockIndexService, *colly.Collector)
		wantErr   bool
	}{
		{
			name:    "empty base URL",
			baseURL: "",
			setupMock: func(ms *storage.MockStorage, ml *logger.MockLogger, mis *storage.MockIndexService, c *colly.Collector) {
				// No setup needed for empty URL test
			},
			wantErr: true,
		},
		{
			name:    "storage ping failure",
			baseURL: "http://example.com",
			setupMock: func(ms *storage.MockStorage, ml *logger.MockLogger, mis *storage.MockIndexService, c *colly.Collector) {
				ms.On("Ping", mock.Anything).Return(errors.New("ping failed"))
				ml.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			wantErr: true,
		},
		{
			name:    "index ensure failure",
			baseURL: "http://example.com",
			setupMock: func(ms *storage.MockStorage, ml *logger.MockLogger, mis *storage.MockIndexService, c *colly.Collector) {
				ms.On("Ping", mock.Anything).Return(nil)
				ml.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
				mis.On("EnsureIndex", mock.Anything, mock.Anything).Return(errors.New("index creation failed"))
			},
			wantErr: true,
		},
		{
			name:    "successful start",
			baseURL: "http://example.com",
			setupMock: func(ms *storage.MockStorage, ml *logger.MockLogger, mis *storage.MockIndexService, c *colly.Collector) {
				ms.On("Ping", mock.Anything).Return(nil)
				ml.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
				ml.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
				mis.On("EnsureIndex", mock.Anything, mock.Anything).Return(nil)

				// Configure collector to immediately return without actually crawling
				c.OnRequest(func(r *colly.Request) {
					r.Abort()
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			mockLogger := logger.NewMockLogger()
			mockIndexService := &storage.MockIndexService{}
			collector := colly.NewCollector()

			tt.setupMock(mockStorage, mockLogger, mockIndexService, collector)

			c := &crawler.Crawler{
				Storage:      mockStorage,
				Logger:       mockLogger,
				IndexService: mockIndexService,
				Config:       &config.Config{},
				Collector:    collector,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := c.Start(ctx, tt.baseURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCrawler_Start_ContextCancellation(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	mockLogger := logger.NewMockLogger()
	mockIndexService := &storage.MockIndexService{}

	mockStorage.On("Ping", mock.Anything).Return(nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockIndexService.On("EnsureIndex", mock.Anything, mock.Anything).Return(nil)

	c := &crawler.Crawler{
		Storage:      mockStorage,
		Logger:       mockLogger,
		IndexService: mockIndexService,
		Config:       &config.Config{},
		Collector:    colly.NewCollector(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := c.Start(ctx, "http://example.com")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
