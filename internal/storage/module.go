package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides the storage as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("storage",
	fx.Provide(
		NewStorage,
	),
)

// NewStorage initializes a new Storage instance
func NewStorage(cfg *config.Config, log logger.Interface) (Interface, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.Elasticsearch.URL == "" {
		return nil, fmt.Errorf("elasticsearch URL is required")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	// Create Elasticsearch config
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
		Transport: transport,
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
		APIKey:    cfg.Elasticsearch.APIKey,
	}

	// Create Elasticsearch client
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %w", err)
	}

	// Create storage instance
	storageInstance := &ElasticsearchStorage{
		ESClient: client,
		Logger:   log,
	}

	// Test connection to Elasticsearch
	if err := storageInstance.TestConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	log.Info("Successfully connected to Elasticsearch",
		"url", cfg.Elasticsearch.URL,
		"using_api_key", cfg.Elasticsearch.APIKey, // != "",
	)

	return storageInstance, nil
}
