package crawler_test

import (
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
