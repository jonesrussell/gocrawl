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
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

const (
	// DefaultMaxRetries is the default number of times to retry failed requests
	DefaultMaxRetries = 3
)

// Module provides the storage module for dependency injection.
var Module = fx.Module("storage",
	fx.Provide(
		func(cfg config.Interface, logger common.Logger) ModuleOptions {
			return ModuleOptions{
				Config:         cfg,
				Logger:         logger,
				Addresses:      cfg.GetElasticsearchConfig().Addresses,
				APIKey:         cfg.GetElasticsearchConfig().APIKey,
				Username:       cfg.GetElasticsearchConfig().Username,
				Password:       cfg.GetElasticsearchConfig().Password,
				Transport:      http.DefaultTransport.(*http.Transport),
				IndexName:      cfg.GetElasticsearchConfig().IndexName,
				ScrollDuration: 5 * time.Minute,
			}
		},
		ProvideElasticsearchClient,
		NewOptionsFromConfig,
		fx.Annotate(
			NewStorage,
			fx.As(new(types.Interface)),
		),
		ProvideIndexManager,
	),
)

// ModuleOptions defines the options for the storage module
type ModuleOptions struct {
	Config         config.Interface
	Logger         common.Logger
	Addresses      []string
	APIKey         string
	Username       string
	Password       string
	Transport      *http.Transport
	IndexName      string
	ScrollDuration time.Duration
}

// ProvideIndexManager creates and returns an IndexManager implementation
func ProvideIndexManager(
	client *es.Client,
	logger common.Logger,
) api.IndexManager {
	return NewIndexManager(client, logger)
}

// ProvideElasticsearchClient provides the Elasticsearch client
func ProvideElasticsearchClient(opts ModuleOptions, log common.Logger) (*es.Client, error) {
	if len(opts.Addresses) == 0 {
		return nil, errors.New("elasticsearch addresses are required")
	}

	log.Debug("Elasticsearch configuration",
		"addresses", opts.Addresses,
		"hasAPIKey", opts.APIKey != "")

	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("failed to get default transport")
	}

	clonedTransport := transport.Clone()

	// Check if TLS verification should be skipped based on environment variable
	skipTLS := false
	if skipTLSStr := os.Getenv("ELASTIC_SKIP_TLS"); skipTLSStr != "" {
		var err error
		skipTLS, err = strconv.ParseBool(skipTLSStr)
		if err != nil {
			log.Warn("Invalid ELASTIC_SKIP_TLS value, defaulting to false", "value", skipTLSStr)
		}
	}

	if skipTLS {
		// #nosec G402 -- InsecureSkipVerify is controlled by ELASTIC_SKIP_TLS environment variable
		clonedTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		log.Warn("TLS certificate verification is disabled")
	}

	cfg := es.Config{
		Addresses: opts.Addresses,
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

	// Configure API key authentication
	if opts.APIKey != "" {
		cfg.APIKey = opts.APIKey
		log.Debug("Using API key authentication")
	} else {
		return nil, errors.New("API key authentication is required")
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
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %s", res.String())
	}

	log.Info("Successfully connected to Elasticsearch", "addresses", opts.Addresses)
	return client, nil
}
