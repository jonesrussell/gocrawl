package app

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/zap"
)

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient() (*elasticsearch.Client, error) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"}, // Consider making this configurable
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}
	return esClient, nil
}

// runCrawler is the function that will be invoked to start the crawling process.
func runCrawler(ctx context.Context) error {
	log := logger.FromContext(ctx) // Get the logger from context
	log.Debug("Starting the crawler...")

	// Example of using the context
	select {
	case <-ctx.Done():
		log.Error("Crawler stopped due to context cancellation", zap.Error(ctx.Err()))
		return ctx.Err()
	default:
		// Continue with the crawling process
	}

	return nil
}

type App struct {
	Storage storage.Interface
}

func NewApp(esClient *elasticsearch.Client, log logger.Interface) (*App, error) {
	storage, err := storage.NewStorage(esClient, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	return &App{
		Storage: storage,
	}, nil
}
