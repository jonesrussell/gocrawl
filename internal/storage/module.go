// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"crypto/x509"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	esconfig "github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
)

const (
	// DefaultMaxRetries is the default number of retries for index operations.
	DefaultMaxRetries = 3
	// DefaultScrollDuration is the default duration for scroll operations.
	DefaultScrollDuration = 5 * time.Minute
	// DefaultResponseHeaderTimeout is the default timeout for response headers.
	DefaultResponseHeaderTimeout = 10 * time.Second
	// DefaultTLSHandshakeTimeout is the default timeout for TLS handshake.
	DefaultTLSHandshakeTimeout = 10 * time.Second
	// DefaultIdleConnTimeout is the default timeout for idle connections.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultMaxIdleConnsPerHost is the default maximum number of idle connections per host.
	DefaultMaxIdleConnsPerHost = 10
	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 100
	// DefaultDialTimeout is the default timeout for dial operations.
	DefaultDialTimeout = 30 * time.Second
	// DefaultDialKeepAlive is the default keep-alive duration for dial operations.
	DefaultDialKeepAlive = 30 * time.Second
	// DefaultExpectContinueTimeout is the default timeout for expect-continue.
	DefaultExpectContinueTimeout = 1 * time.Second
)

// createTLSConfig creates a TLS configuration with appropriate security settings
func createTLSConfig(esConfig *esconfig.Config) (*tls.Config, error) {
	// Create basic TLS config with minimum version 1.2
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Handle CA certificates if provided
	if esConfig.TLS != nil && esConfig.TLS.CAFile != "" {
		caCert, err := os.ReadFile(esConfig.TLS.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Handle client certificates if provided
	if esConfig.TLS != nil && esConfig.TLS.CertFile != "" && esConfig.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(esConfig.TLS.CertFile, esConfig.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Handle insecure skip verify if configured
	if esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return tlsConfig, nil
}

// createTransport creates a configured HTTP transport for Elasticsearch
func createTransport(esConfig *esconfig.Config, logger logger.Interface) (*http.Transport, error) {
	logger.Debug("Creating HTTP transport")

	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("failed to get default transport")
	}

	clonedTransport := transport.Clone()

	// Create and set TLS config
	tlsConfig, err := createTLSConfig(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
	}
	clonedTransport.TLSClientConfig = tlsConfig

	// Set timeouts
	clonedTransport.ResponseHeaderTimeout = DefaultResponseHeaderTimeout
	clonedTransport.ExpectContinueTimeout = 1 * time.Second
	clonedTransport.TLSHandshakeTimeout = DefaultTLSHandshakeTimeout
	clonedTransport.IdleConnTimeout = DefaultIdleConnTimeout

	// Enable HTTP/2 support
	if configErr := configureTransport(clonedTransport); configErr != nil {
		logger.Warn("Failed to enable HTTP/2 support", "error", configErr)
	}

	logger.Debug("Created HTTP transport with TLS configuration",
		"responseHeaderTimeout", clonedTransport.ResponseHeaderTimeout,
		"expectContinueTimeout", clonedTransport.ExpectContinueTimeout,
		"tlsHandshakeTimeout", clonedTransport.TLSHandshakeTimeout,
		"idleConnTimeout", clonedTransport.IdleConnTimeout,
		"tlsInsecureSkipVerify", tlsConfig.InsecureSkipVerify)

	return clonedTransport, nil
}

// configureTransport configures the HTTP transport with appropriate timeouts and settings
func configureTransport(transport *http.Transport) error {
	transport.ResponseHeaderTimeout = DefaultResponseHeaderTimeout
	transport.TLSHandshakeTimeout = DefaultTLSHandshakeTimeout
	transport.IdleConnTimeout = DefaultIdleConnTimeout

	// Configure HTTP/2
	if configErr := http2.ConfigureTransport(transport); configErr != nil {
		return fmt.Errorf("failed to configure HTTP/2 transport: %w", configErr)
	}

	return nil
}

// createClientConfig creates an Elasticsearch client configuration with the provided settings.
func createClientConfig(
	esConfig *esconfig.Config,
	transport *http.Transport,
	logger logger.Interface,
) elasticsearch.Config {
	// Log configuration details
	logger.Debug("Creating Elasticsearch client configuration",
		"addresses", esConfig.Addresses,
		"hasAPIKey", esConfig.APIKey != "",
		"hasUsername", esConfig.Username != "",
		"hasPassword", esConfig.Password != "",
		"tlsInsecureSkipVerify", esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify,
		"tlsHasCAFile", esConfig.TLS != nil && esConfig.TLS.CAFile != "",
		"tlsHasCertFile", esConfig.TLS != nil && esConfig.TLS.CertFile != "",
		"tlsHasKeyFile", esConfig.TLS != nil && esConfig.TLS.KeyFile != "",
		"retryEnabled", esConfig.Retry.Enabled,
		"maxRetries", esConfig.Retry.MaxRetries)

	// Create client configuration
	cfg := elasticsearch.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		APIKey:    esConfig.APIKey,
		Transport: transport,
		RetryOnStatus: []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		MaxRetries:            esConfig.Retry.MaxRetries,
		DiscoverNodesOnStart:  esConfig.DiscoverNodes,
		DiscoverNodesInterval: 0, // Disable periodic node discovery
		CompressRequestBody:   true,
	}

	// Log final configuration (excluding sensitive fields)
	logger.Debug("Final Elasticsearch client configuration",
		"addresses", cfg.Addresses,
		"hasAPIKey", cfg.APIKey != "",
		"enableMetrics", cfg.EnableMetrics,
		"enableDebugLogger", cfg.EnableDebugLogger,
		"enableCompatibilityMode", cfg.EnableCompatibilityMode,
		"compressRequestBody", cfg.CompressRequestBody,
		"disableRetry", cfg.DisableRetry,
		"retryOnStatus", cfg.RetryOnStatus,
		"maxRetries", cfg.MaxRetries,
		"discoverNodesOnStart", cfg.DiscoverNodesOnStart,
		"discoverNodesInterval", cfg.DiscoverNodesInterval,
		"transport", map[string]any{
			"tlsInsecureSkipVerify": transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify,
			"hasRootCAs":            transport.TLSClientConfig != nil && transport.TLSClientConfig.RootCAs != nil,
			"hasCertificates":       transport.TLSClientConfig != nil && len(transport.TLSClientConfig.Certificates) > 0,
			"minVersion": func() uint16 {
				if transport.TLSClientConfig != nil {
					return transport.TLSClientConfig.MinVersion
				}
				return 0
			}(),
		})

	return cfg
}

// NewElasticsearchClient creates a new Elasticsearch client with the provided configuration
func NewElasticsearchClient(cfg config.Interface, logger logger.Interface) (*elasticsearch.Client, error) {
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
	transport, err := createTransport(esConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create client config
	esCfg := createClientConfig(esConfig, transport, logger)

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
				transport := &http.Transport{
					MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
					ResponseHeaderTimeout: DefaultResponseHeaderTimeout,
					DialContext: (&net.Dialer{
						Timeout:   DefaultDialTimeout,
						KeepAlive: DefaultDialKeepAlive,
					}).DialContext,
					MaxIdleConns:          DefaultMaxIdleConns,
					IdleConnTimeout:       DefaultIdleConnTimeout,
					TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
					ExpectContinueTimeout: DefaultExpectContinueTimeout,
				}

				// Configure TLS if enabled
				if esConfig.TLS != nil {
					// InsecureSkipVerify is used for development/testing environments only
					// and should be disabled in production. This is a security risk.
					transport.TLSClientConfig = &tls.Config{
						InsecureSkipVerify: esConfig.TLS.InsecureSkipVerify, // #nosec G402 - This is configurable and documented
					}
				}

				// Create Elasticsearch config
				config := elasticsearch.Config{
					Addresses: esConfig.Addresses,
					Username:  esConfig.Username,
					Password:  esConfig.Password,
					APIKey:    esConfig.APIKey,
					Transport: transport,
				}

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

		// Provide the storage
		fx.Annotate(
			func(
				client *elasticsearch.Client,
				logger logger.Interface,
				cfg config.Interface,
			) types.Interface {
				opts := DefaultOptions()
				opts.Addresses = cfg.GetElasticsearchConfig().Addresses
				opts.Username = cfg.GetElasticsearchConfig().Username
				opts.Password = cfg.GetElasticsearchConfig().Password
				opts.APIKey = cfg.GetElasticsearchConfig().APIKey
				opts.IndexName = cfg.GetElasticsearchConfig().IndexName
				return NewStorage(client, logger, opts)
			},
			fx.As(new(types.Interface)),
		),
	),
)
