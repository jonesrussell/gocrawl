package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/viper"
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
func NewStorage(esClient *elasticsearch.Client) (Interface, error) {
	if esClient == nil {
		return nil, errors.New("elasticsearch client is nil")
	}

	// Log the Elasticsearch client information for debugging
	fmt.Printf("Elasticsearch client initialized: %+v\n", esClient)

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
	if err := storageInstance.TestConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	return storageInstance, nil
}

// Provide the Elasticsearch client as a dependency
func ProvideElasticsearchClient() (*elasticsearch.Client, error) {
	// Create a custom HTTP transport that skips TLS verification
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200", // Ensure this matches your Elasticsearch URL
		},
		Transport: transport,                           // Use the custom transport
		Username:  viper.GetString("ELASTIC_USERNAME"), // Get username from config
		Password:  viper.GetString("ELASTIC_PASSWORD"), // Get password from config
		APIKey:    viper.GetString("ELASTIC_API_KEY"),  // Get API key from config
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Log the configuration for debugging
	fmt.Printf("Connecting to Elasticsearch at %s\n", cfg.Addresses[0])

	return client, nil
}
