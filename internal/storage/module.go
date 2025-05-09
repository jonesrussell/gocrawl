// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	esconfig "github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// DefaultMaxRetries is the default maximum number of retries for failed requests.
const DefaultMaxRetries = 3

// DefaultRetryDelay is the default delay between retries.
const DefaultRetryDelay = 1 * time.Second

// DefaultMaxBodySize is the default maximum body size for responses.
const DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB

// DefaultRequestTimeout is the default timeout for requests.
const DefaultRequestTimeout = 30 * time.Second

// DefaultScrollDuration is the default duration for scroll operations.
const DefaultScrollDuration = 5 * time.Minute

// DefaultDialTimeout is the default timeout for dial operations.
const DefaultDialTimeout = 30 * time.Second

// DefaultDialKeepAlive is the default keep-alive duration for dial operations.
const DefaultDialKeepAlive = 30 * time.Second

// DefaultMaxIdleConnsPerHost is the default maximum number of idle connections per host.
const DefaultMaxIdleConnsPerHost = 10

// DefaultResponseHeaderTimeout is the default timeout for response headers.
const DefaultResponseHeaderTimeout = 5 * time.Second

// CreateTransport creates a new HTTP transport with appropriate settings
func CreateTransport(esConfig *esconfig.Config) *http.Transport {
	transport := &http.Transport{
		MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
		ResponseHeaderTimeout: DefaultResponseHeaderTimeout,
	}

	// Configure TLS if enabled
	if esConfig.TLS != nil && esConfig.TLS.Enabled {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: esConfig.TLS.InsecureSkipVerify, // #nosec G402 - This is configurable and documented
		}
	}

	return transport
}

// CreateClientConfig creates an Elasticsearch client configuration with the provided settings.
func CreateClientConfig(
	esConfig *esconfig.Config,
	transport *http.Transport,
) es.Config {
	return es.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		APIKey:    esConfig.APIKey,
		Transport: transport,
	}
}

// ClientParams contains dependencies for creating an Elasticsearch client
type ClientParams struct {
	fx.In

	Config config.Interface
}

// ClientResult contains the Elasticsearch client
type ClientResult struct {
	fx.Out

	Client *es.Client
}

// NewClient creates a new Elasticsearch client
func NewClient(p ClientParams) (ClientResult, error) {
	esConfig := p.Config.GetElasticsearchConfig()
	if esConfig == nil {
		return ClientResult{}, errors.New("elasticsearch configuration is required")
	}

	transport := CreateTransport(esConfig)
	clientConfig := CreateClientConfig(esConfig, transport)

	client, err := es.NewClient(clientConfig)
	if err != nil {
		return ClientResult{}, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	res, err := client.Ping()
	if err != nil {
		return ClientResult{}, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return ClientResult{}, fmt.Errorf("error pinging Elasticsearch: %s", res.String())
	}

	return ClientResult{Client: client}, nil
}

// StorageParams contains dependencies for creating a storage implementation
type StorageParams struct {
	fx.In

	Client *es.Client
	Logger logger.Interface
}

// StorageResult contains the storage implementation
type StorageResult struct {
	fx.Out

	Storage types.Interface
}

// NewStorage creates a new storage implementation
func NewStorage(p StorageParams) StorageResult {
	opts := DefaultOptions()
	storage := &Storage{
		client: p.Client,
		logger: p.Logger,
		opts:   opts,
	}
	storage.indexManager = NewElasticsearchIndexManager(p.Client, p.Logger)
	return StorageResult{Storage: storage}
}

// Module provides all storage-related dependencies
var Module = fx.Module("storage",
	fx.Provide(
		NewClient,
		NewStorage,
	),
)
