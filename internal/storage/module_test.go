package storage

import (
	"net/http"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// Initialize testConfig
var testConfig = &config.Config{
	Elasticsearch: config.ElasticsearchConfig{
		URL: "http://localhost:9200", // or use a test URL
	},
	Crawler: config.CrawlerConfig{
		Transport: http.DefaultTransport,
	},
}

func TestModule(t *testing.T) {
	t.Run("module provides storage", func(t *testing.T) {
		app := fxtest.New(t,
			Module,
			fx.Provide(
				func() *config.Config {
					return testConfig
				},
				func() logger.Interface {
					return logger.NewMockCustomLogger()
				},
			),
		)
		assert.NoError(t, app.Err())
	})
}
