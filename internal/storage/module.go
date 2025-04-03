package storage

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Constants for timeouts and retries
const (
	DefaultMaxRetries     = 3
	DefaultScrollDuration = 5 * time.Minute
)

// NewElasticsearchClient creates a new Elasticsearch client with the provided configuration
func NewElasticsearchClient(cfg config.Interface, logger logger.Interface) (*es.Client, error) {
	esConfig := cfg.GetElasticsearchConfig()
	if len(esConfig.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	logger.Debug("Elasticsearch configuration",
		"addresses", esConfig.Addresses,
		"hasAPIKey", esConfig.APIKey != "",
		"hasBasicAuth", esConfig.Username != "" && esConfig.Password != "")

	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("failed to get default transport")
	}

	clonedTransport := transport.Clone()

	// Check if TLS verification should be skipped based on environment variable
	skipTLS := false
	switch skipTLSStr := os.Getenv("ELASTIC_SKIP_TLS"); {
	case skipTLSStr == "":
		// Default to false
	case skipTLSStr != "":
		var err error
		skipTLS, err = strconv.ParseBool(skipTLSStr)
		if err != nil {
			logger.Warn("Invalid ELASTIC_SKIP_TLS value, defaulting to false", "value", skipTLSStr)
		}
	}

	if skipTLS {
		// #nosec G402 -- InsecureSkipVerify is controlled by ELASTIC_SKIP_TLS environment variable
		clonedTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		logger.Warn("TLS certificate verification is disabled")
	}

	esCfg := es.Config{
		Addresses: esConfig.Addresses,
		Transport: clonedTransport,
		// Client configuration
		EnableMetrics:       false,
		EnableDebugLogger:   false,
		CompressRequestBody: true,
		DisableRetry:        false,
		RetryOnStatus:       []int{502, 503, 504},
		MaxRetries:          DefaultMaxRetries,
		RetryBackoff:        func(i int) time.Duration { return time.Duration(i) * 100 * time.Millisecond },
	}

	// Configure authentication
	switch {
	case esConfig.APIKey != "":
		esCfg.APIKey = esConfig.APIKey
		logger.Debug("Using API key authentication")
	case esConfig.Username != "" && esConfig.Password != "":
		esCfg.Username = esConfig.Username
		esCfg.Password = esConfig.Password
		logger.Debug("Using basic authentication")
	default:
		return nil, errors.New("either API key or basic authentication is required")
	}

	client, err := es.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test the connection
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
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %s", res.String())
	}

	logger.Info("Successfully connected to Elasticsearch", "addresses", esConfig.Addresses)
	return client, nil
}

// Module provides the storage dependencies.
var Module = fx.Module("storage",
	fx.Provide(
		NewOptionsFromConfig,
		NewElasticsearchClient,
		NewElasticsearchIndexManager,
		NewStorage,
		// Provide SearchManager implementation
		func(storage Interface) interfaces.SearchManager {
			return &searchManagerWrapper{storage}
		},
	),
)

// searchManagerWrapper wraps storage.Interface to implement interfaces.SearchManager
type searchManagerWrapper struct {
	Interface
}

func (w *searchManagerWrapper) Close() error {
	return nil
}
