package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides storage dependencies
var Module = fx.Module("storage",
	fx.Provide(
		NewOptionsFromConfig,
		ProvideElasticsearchClient,
		NewElasticsearchStorage,
	),
)

// NewElasticsearchStorage creates a new ElasticsearchStorage instance
func NewElasticsearchStorage(
	client *elasticsearch.Client,
	logger logger.Interface,
	opts Options,
) Interface {
	return &ElasticsearchStorage{
		ESClient: client,
		Logger:   logger,
		opts:     opts,
	}
}

// ProvideElasticsearchClient provides the Elasticsearch client
func ProvideElasticsearchClient(opts Options, log logger.Interface) (*elasticsearch.Client, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{opts.URL},
		Username:  opts.Username,
		Password:  opts.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test the connection
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New("failed to connect to Elasticsearch")
	}

	return client, nil
}

// NewStorage initializes a new Storage instance
func NewStorage(esClient *elasticsearch.Client, opts Options, log logger.Interface) (Interface, error) {
	if esClient == nil {
		return nil, errors.New("elasticsearch client is nil")
	}

	// Log the Elasticsearch client information for debugging
	log.Info("Elasticsearch client initialized", "client", esClient)

	// Attempt to ping Elasticsearch to check connectivity
	res, err := esClient.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	storageInstance := &ElasticsearchStorage{
		ESClient: esClient,
		Logger:   log,
		opts:     opts,
	}

	// Test connection to Elasticsearch
	if testErr := storageInstance.TestConnection(context.Background()); testErr != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", testErr)
	}

	return storageInstance, nil
}
