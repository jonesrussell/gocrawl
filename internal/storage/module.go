// Package storage implements the storage layer for the application.
package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Constants for timeouts and retries
const (
	DefaultMaxRetries     = 3
	DefaultScrollDuration = 5 * time.Minute
)

// createTLSConfig creates a TLS configuration with appropriate security settings
func createTLSConfig(esConfig *elasticsearch.Config, logger logger.Interface) *tls.Config {
	if esConfig.TLS == nil || !esConfig.TLS.Enabled {
		return nil
	}

	// InsecureSkipVerify is only used in development environments
	// and should be disabled in production. This is a security risk
	// and should only be used for testing or development purposes.
	// See https://pkg.go.dev/crypto/tls#Config for more information.
	// SECURITY WARNING: Setting InsecureSkipVerify to true means that
	// the client will accept any certificate presented by the server,
	// making the connection vulnerable to man-in-the-middle attacks.
	// This should never be used in production environments.
	if esConfig.TLS.InsecureSkipVerify {
		logger.Warn(
			"TLS certificate verification is disabled - " +
				"this is a security risk and should never be used in production",
		)
	}

	// #nosec G402 - InsecureSkipVerify is configurable and documented
	return &tls.Config{
		InsecureSkipVerify: esConfig.TLS.InsecureSkipVerify,
	}
}

// createTransport creates a configured HTTP transport for Elasticsearch
func createTransport(esConfig *elasticsearch.Config, logger logger.Interface) (*http.Transport, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("failed to get default transport")
	}

	clonedTransport := transport.Clone()
	clonedTransport.TLSClientConfig = createTLSConfig(esConfig, logger)

	return clonedTransport, nil
}

// createClientConfig creates an Elasticsearch client configuration
func createClientConfig(esConfig *elasticsearch.Config, transport *http.Transport, logger logger.Interface) es.Config {
	// Log detailed configuration information
	logger.Debug("Creating Elasticsearch client configuration",
		"addresses", esConfig.Addresses,
		"hasAPIKey", esConfig.APIKey != "",
		"hasUsername", esConfig.Username != "",
		"hasPassword", esConfig.Password != "",
		"tlsEnabled", esConfig.TLS != nil && esConfig.TLS.Enabled,
		"tlsInsecureSkipVerify", esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify)

	// Create the client configuration
	cfg := es.Config{
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
		DiscoverNodesInterval: 0, // Set to 0 to disable periodic node discovery
	}

	// Log the final configuration (excluding sensitive fields)
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
		"discoverNodesInterval", cfg.DiscoverNodesInterval)

	return cfg
}

// NewElasticsearchClient creates a new Elasticsearch client with the provided configuration
func NewElasticsearchClient(cfg config.Interface, logger logger.Interface) (*es.Client, error) {
	esConfig := cfg.GetElasticsearchConfig()
	if len(esConfig.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	if esConfig.APIKey == "" {
		return nil, errors.New("API key is required")
	}

	// Log detailed configuration information
	logger.Debug("Elasticsearch configuration",
		"addresses", esConfig.Addresses,
		"hasAPIKey", esConfig.APIKey != "",
		"hasBasicAuth", esConfig.Username != "" && esConfig.Password != "",
		"tls.enabled", esConfig.TLS != nil && esConfig.TLS.Enabled,
		"tls.insecure_skip_verify", esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify)

	// Create transport
	transport, err := createTransport(esConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create client config
	esCfg := createClientConfig(esConfig, transport, logger)

	// Create client
	client, err := es.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		// Log the full error response
		body, _ := io.ReadAll(res.Body)
		logger.Error("Elasticsearch ping error",
			"status", res.StatusCode,
			"error", string(body))
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %s", res.String())
	}

	logger.Info("Successfully connected to Elasticsearch", "addresses", esConfig.Addresses)
	return client, nil
}

// Module provides the storage module
var Module = fx.Module("storage",
	fx.Provide(
		// Provide Elasticsearch client
		NewElasticsearchClient,
		// Provide storage options
		func(cfg config.Interface) Options {
			return Options{
				IndexName: cfg.GetElasticsearchConfig().IndexName,
			}
		},
		// Provide storage client
		NewStorage,
		// Provide search manager
		NewSearchManager,
	),
)
