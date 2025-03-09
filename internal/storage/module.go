package storage

import (
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/api"
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
		fx.Annotate(
			func(s Interface) (api.SearchManager, error) {
				sm, ok := s.(api.SearchManager)
				if !ok {
					return nil, errors.New("storage implementation does not satisfy api.SearchManager interface")
				}
				return sm, nil
			},
			fx.As(new(api.SearchManager)),
		),
	),
)

// ProvideIndexManager creates and returns an IndexManager implementation
func ProvideIndexManager(client *es.Client, log logger.Interface) (api.IndexManager, error) {
	if client == nil {
		return nil, errors.New("elasticsearch client is required")
	}
	return NewElasticsearchIndexManager(client, log), nil
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

	log.Debug("Elasticsearch configuration",
		"addresses", opts.Addresses,
		"hasUsername", opts.Username != "",
		"hasPassword", opts.Password != "",
		"hasAPIKey", opts.APIKey != "",
		"skipTLS", opts.SkipTLS)

	cfg := es.Config{
		Addresses: opts.Addresses,
	}

	// Configure authentication - prefer API key over username/password
	if opts.APIKey != "" {
		cfg.APIKey = opts.APIKey
		log.Debug("Using API key authentication")
	} else if opts.Username != "" && opts.Password != "" {
		cfg.Username = opts.Username
		cfg.Password = opts.Password
		log.Debug("Using username/password authentication")
	} else {
		log.Debug("No authentication credentials provided")
	}

	// Configure TLS if needed
	if opts.Transport != nil {
		cfg.Transport = opts.Transport
		log.Debug("Using custom transport with TLS configuration")
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
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %s", res.String())
	}

	log.Info("Successfully connected to Elasticsearch", "addresses", opts.Addresses)
	return client, nil
}
