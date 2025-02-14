package app

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config) (logger.Interface, error) {
	log, err := logger.NewDevelopmentLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return log, nil
}

func runCrawler(ctx context.Context, crawler *crawler.Crawler) error {
	// Start the crawler
	if err := crawler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Wait for the crawler to finish (if necessary)
	// You might want to implement a wait group or similar mechanism here

	return nil
}

type App struct {
	Storage storage.Interface
}

func NewApp(esClient *elasticsearch.Client) (*App, error) {
	storage, err := storage.NewStorage(esClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	return &App{
		Storage: storage,
	}, nil
}
