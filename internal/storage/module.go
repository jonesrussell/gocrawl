// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
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

// createTransport creates a new HTTP transport with appropriate settings
func createTransport(esConfig *esconfig.Config) *http.Transport {
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

// createClientConfig creates an Elasticsearch client configuration with the provided settings.
func createClientConfig(
	esConfig *esconfig.Config,
	transport *http.Transport,
) elasticsearch.Config {
	return elasticsearch.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		APIKey:    esConfig.APIKey,
		Transport: transport,
	}
}

// NewElasticsearchClient creates a new Elasticsearch client with the provided configuration
func NewElasticsearchClient(cfg config.Interface, logger logger.Interface) (*elasticsearch.Client, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	// Get Elasticsearch configuration
	esConfig := cfg.GetElasticsearchConfig()
	if esConfig == nil {
		return nil, errors.New("elasticsearch config is nil")
	}

	if len(esConfig.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	// Create transport with TLS configuration
	transport := createTransport(esConfig)

	// Create client config
	esCfg := createClientConfig(esConfig, transport)

	// Create client
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error pinging Elasticsearch: %s", res.String())
	}

	return client, nil
}

// Module provides the storage module for dependency injection.
var Module = fx.Module("storage",
	fx.Provide(
		// Provide the Elasticsearch client
		NewElasticsearchClient,

		// Provide the index manager
		NewElasticsearchIndexManager,

		// Provide the storage interface
		func(client *elasticsearch.Client, logger logger.Interface) types.Interface {
			defaultOpts := DefaultOptions()
			return NewStorage(client, logger, &defaultOpts)
		},

		// Provide the search manager
		NewSearchManager,
	),
)
