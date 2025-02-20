package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the storage module and its dependencies
var Module = fx.Module("storage",
	fx.Provide(
		ProvideElasticsearchClient,
		NewStorage,       // Ensure this is provided
		NewSearchService, // Provide the search service
	),
)

// NewStorage initializes a new Storage instance
func NewStorage(esClient *elasticsearch.Client, log logger.Interface) (Interface, error) {
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

	// Create storage instance
	storageInstance := &ElasticsearchStorage{
		ESClient: esClient,
		Logger:   log,
	}

	// Test connection to Elasticsearch
	if testErr := storageInstance.TestConnection(context.Background()); testErr != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", testErr)
	}

	return storageInstance, nil
}

// ProvideElasticsearchClient initializes the Elasticsearch client
func ProvideElasticsearchClient(cfg *config.Config, log logger.Interface) (*elasticsearch.Client, error) {
	// Create a custom HTTP transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			//nolint:gosec // We are using the SkipTLS setting from the config
			InsecureSkipVerify: cfg.Elasticsearch.SkipTLS, // Use the SkipTLS setting from the config
		},
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
		APIKey:    cfg.Elasticsearch.APIKey,
		Transport: transport, // Use the custom transport
	})
	if err != nil {
		log.Error("Failed to create Elasticsearch client", "error", err)
		return nil, err
	}
	log.Info("Elasticsearch client initialized successfully")
	return client, nil
}
