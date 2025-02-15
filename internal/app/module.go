package app

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the application module and its dependencies
var Module = fx.Module("app",
	fx.Provide(
		logger.NewLogger,
		NewElasticsearchClient, // Provide the Elasticsearch client
		func() func(ctx context.Context) error {
			return runCrawler // Ensure runCrawler is provided correctly
		},
	),
	fx.Invoke(
		func(ctx context.Context, esClient *elasticsearch.Client) error {
			log := logger.FromContext(ctx)
			log.Debug("Invoking runCrawler with provided context")
			return runCrawler(ctx)
		},
	),
)
