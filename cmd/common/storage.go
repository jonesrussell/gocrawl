package common

import (
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// StorageResult holds both storage interface and index manager.
// This consolidates the common pattern used across all commands.
type StorageResult struct {
	Storage      types.Interface
	IndexManager types.IndexManager
}

// CreateStorageClient creates an Elasticsearch client with the given config and logger.
// This consolidates the duplicate createStorageClientFor* functions.
func CreateStorageClient(cfg config.Interface, log logger.Interface) (*es.Client, error) {
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: cfg,
		Logger: log,
	})
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	return clientResult.Client, nil
}

// CreateStorage creates both storage client and storage instance in one call.
// This consolidates the common pattern used across all commands.
func CreateStorage(cfg config.Interface, log logger.Interface) (*StorageResult, error) {
	// Create storage client
	client, err := CreateStorageClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}

	// Create storage
	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: client,
	})
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}

	return &StorageResult{
		Storage:      storageResult.Storage,
		IndexManager: storageResult.IndexManager,
	}, nil
}
