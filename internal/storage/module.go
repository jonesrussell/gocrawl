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

type ElasticsearchClient struct {
	Client *elasticsearch.Client
}

// Module provides the storage module and its dependencies
var Module = fx.Module("storage",
	fx.Provide(
		ProvideElasticsearchClient, // Function to provide an Elasticsearch client
		NewStorage,                 // Function to create a new storage instance
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
	}

	// Test connection to Elasticsearch
	if testErr := storageInstance.TestConnection(context.Background()); testErr != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", testErr)
	}

	return storageInstance, nil
}

// Provide the Elasticsearch client as a dependency
func ProvideElasticsearchClient(log logger.Interface, cfg *config.Config) (*elasticsearch.Client, error) {
	// Create a custom HTTP transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			//nolint:gosec // Allow insecure skip verify only in development
			InsecureSkipVerify: cfg.Crawler.SkipTLS, // Use the SkipTLS setting from the config
		},
	}

	cfgElasticsearch := elasticsearch.Config{
		Addresses: []string{
			cfg.Elasticsearch.URL, // Use the URL from the config
		},
		Transport: transport,                  // Use the custom transport
		Username:  cfg.Elasticsearch.Username, // Get username from config
		Password:  cfg.Elasticsearch.Password, // Get password from config
		APIKey:    cfg.Elasticsearch.APIKey,   // Get API key from config
	}

	client, err := elasticsearch.NewClient(cfgElasticsearch)
	if err != nil {
		return nil, err
	}

	// Log the configuration for debugging
	log.Info("Connecting to Elasticsearch", "address", cfgElasticsearch.Addresses[0])

	return client, nil
}
