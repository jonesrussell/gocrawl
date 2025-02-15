package crawler_test

import (
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

func TestCrawler_Stop(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	c := &crawler.Crawler{
		Logger: mockLogger,
	}

	// There is no specific behavior to test for Stop method since it does nothing
	c.Stop()
	// Here we could test if any cleanup code is executed
}

func TestCrawler_ProcessPage(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := storage.NewMockStorage()
	mockIndexService := storage.NewMockIndexService()
	mockArticleService := article.NewMockService()

	// Ensure the config is properly initialized
	configInstance := &config.Config{
		Crawler: config.CrawlerConfig{
			BaseURL:   "http://example.com",
			MaxDepth:  3,
			RateLimit: time.Second,
		},
	}

	mockHTMLElement := &colly.HTMLElement{}

	mockArticle := &models.Article{ID: "1", Title: "Test Article"}
	mockArticleService.On("ExtractArticle", mockHTMLElement).Return(mockArticle)
	mockLogger.On("Debug", "Processing page", "url", mock.Anything).Return()
	mockLogger.On("Debug", "Article extracted", "url", mock.Anything, "title", "Test Article").Return()
	mockStorage.On("IndexDocument", mock.Anything, "articles", "1", mockArticle).Return(nil)

	c := &crawler.Crawler{
		Storage:        mockStorage,
		Logger:         mockLogger,
		ArticleService: mockArticleService,
		IndexSvc:       mockIndexService,
		Config:         configInstance, // Ensure this is set correctly
	}

	// Call ProcessPage with the mock HTML element
	c.ProcessPage(mockHTMLElement)

	// Assert expectations
	mockArticleService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
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

func TestCrawler_GetMaxDepth(t *testing.T) {
	configInstance := &config.Config{
		Crawler: config.CrawlerConfig{
			MaxDepth: 3,
		},
	}
	c := &crawler.Crawler{Config: configInstance}
	assert.Equal(t, 3, c.GetMaxDepth())
}

func TestCrawler_GetRateLimit(t *testing.T) {
	configInstance := &config.Config{
		Crawler: config.CrawlerConfig{
			RateLimit: time.Second,
		},
	}
	c := &crawler.Crawler{Config: configInstance}
	assert.Equal(t, time.Second, c.GetRateLimit())
}
