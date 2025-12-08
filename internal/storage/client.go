// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ClientParams contains dependencies for creating the Elasticsearch client
type ClientParams struct {
	Config config.Interface
	Logger logger.Interface
}

// ClientResult contains the Elasticsearch client
type ClientResult struct {
	Client *es.Client
}

// NewClient creates a new Elasticsearch client
func NewClient(p ClientParams) (ClientResult, error) {
	// Get Elasticsearch config
	esConfig := p.Config.GetElasticsearchConfig()
	if esConfig == nil {
		return ClientResult{}, errors.New("elasticsearch configuration is required")
	}

	// Log the addresses being used for debugging
	if len(esConfig.Addresses) > 0 {
		p.Logger.Debug("Connecting to Elasticsearch", "addresses", esConfig.Addresses)
	}

	// Create transport
	transport := CreateTransport(esConfig)
	clientConfig := CreateClientConfig(esConfig, transport)

	// Create client
	client, err := es.NewClient(*clientConfig)
	if err != nil {
		return ClientResult{}, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Verify client connection
	res, err := client.Ping()
	if err != nil {
		return ClientResult{}, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return ClientResult{}, fmt.Errorf("error pinging Elasticsearch: %s", res.String())
	}

	return ClientResult{
		Client: client,
	}, nil
}

// CreateTransport creates an HTTP transport with TLS configuration.
func CreateTransport(cfg *elasticsearch.Config) *http.Transport {
	transport := &http.Transport{}

	// Configure TLS if enabled
	if cfg.TLS != nil && cfg.TLS.Enabled {
		tlsConfig := &tls.Config{
			//nolint:gosec // InsecureSkipVerify is configurable for development/testing environments
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}

		// Load certificates if provided
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err == nil {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		// Load CA certificate if provided
		// TODO: Implement CA certificate loading using crypto/x509
		if cfg.TLS.CAFile != "" {
			_ = cfg.TLS.CAFile // Placeholder for future CA certificate loading implementation
		}

		transport.TLSClientConfig = tlsConfig
	}

	return transport
}

// CreateClientConfig creates an Elasticsearch client configuration.
func CreateClientConfig(cfg *elasticsearch.Config, transport *http.Transport) *es.Config {
	clientConfig := es.Config{
		Addresses: cfg.Addresses,
		Transport: transport,
	}

	// Configure authentication
	if cfg.APIKey != "" {
		clientConfig.APIKey = cfg.APIKey
	} else if cfg.Username != "" && cfg.Password != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	// Configure cloud settings if provided
	if cfg.Cloud.ID != "" {
		clientConfig.CloudID = cfg.Cloud.ID
	}
	if cfg.Cloud.APIKey != "" {
		clientConfig.APIKey = cfg.Cloud.APIKey
	}

	return &clientConfig
}
