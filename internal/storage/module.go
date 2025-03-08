package storage

import (
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/indexing"
	"github.com/jonesrussell/gocrawl/internal/indexing/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Module provides storage dependencies
var Module = fx.Module("storage",
	fx.Provide(
		NewOptionsFromConfig,
		ProvideElasticsearchClient,
		NewElasticsearchStorage,
		ProvideIndexManager,
	),
)

// ProvideIndexManager creates and returns an IndexManager implementation
func ProvideIndexManager(client *es.Client, log logger.Interface) (indexing.Manager, error) {
	if client == nil {
		return nil, errors.New("elasticsearch client is required")
	}
	return elasticsearch.NewManager(client, log), nil
}

// NewElasticsearchStorage creates a new ElasticsearchStorage instance
func NewElasticsearchStorage(
	client *es.Client,
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
func ProvideElasticsearchClient(opts Options, log logger.Interface) (*es.Client, error) {
	if len(opts.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	cfg := es.Config{
		Addresses: opts.Addresses,
		Username:  opts.Username,
		Password:  opts.Password,
	}

	// Configure TLS if needed
	if opts.SkipTLS {
		cfg.Transport = opts.Transport
	}

	client, err := es.NewClient(cfg)
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

	log.Info("Successfully connected to Elasticsearch", "addresses", opts.Addresses)
	return client, nil
}
