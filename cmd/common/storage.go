package common

import (
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"

	"github.com/jonesrussell/gocrawl/internal/config"

	"github.com/jonesrussell/gocrawl/internal/logger"

	"github.com/jonesrussell/gocrawl/internal/storage"
)

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
