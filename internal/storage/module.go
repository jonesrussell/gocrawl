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
	if esConfig.TLS != nil {
		// InsecureSkipVerify is used for development/testing environments only
		// and should be disabled in production. This is a security risk.
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
	// Get Elasticsearch configuration
	esConfig := cfg.GetElasticsearchConfig()
	if len(esConfig.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	// Log detailed configuration information
	logger.Debug("Elasticsearch configuration",
		"addresses", esConfig.Addresses,
		"hasAPIKey", esConfig.APIKey != "",
		"hasBasicAuth", esConfig.Username != "" && esConfig.Password != "",
		"tls.insecure_skip_verify", esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify,
		"tls.has_ca_file", esConfig.TLS != nil && esConfig.TLS.CAFile != "",
		"tls.has_cert_file", esConfig.TLS != nil && esConfig.TLS.CertFile != "",
		"tls.has_key_file", esConfig.TLS != nil && esConfig.TLS.KeyFile != "")

	// Create transport
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
		fx.Annotate(
			func(logger logger.Interface, cfg config.Interface) (*elasticsearch.Client, error) {
				esConfig := cfg.GetElasticsearchConfig()

				// Create transport with TLS configuration
				transport := createTransport(esConfig)

				// Create Elasticsearch config
				config := createClientConfig(esConfig, transport)

				client, err := elasticsearch.NewClient(config)
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
					return nil, fmt.Errorf("error pinging Elasticsearch: %s", res.String())
				}

				return client, nil
			},
			fx.ResultTags(`name:"elasticsearchClient"`),
		),

		// Provide the index manager
		fx.Annotate(
			NewElasticsearchIndexManager,
			fx.ResultTags(`name:"indexManager"`),
		),

		// Provide the storage interface
		fx.Annotate(
			func(client *elasticsearch.Client, logger logger.Interface) types.Interface {
				defaultOpts := DefaultOptions()
				return NewStorage(client, logger, &defaultOpts)
			},
			fx.ParamTags(`name:"elasticsearchClient"`),
		),

		// Provide the search manager
		fx.Annotate(
			NewSearchManager,
			fx.ParamTags(``, ``),
		),
	),
)
