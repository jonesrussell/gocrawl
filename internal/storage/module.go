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
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

const (
	// DefaultMaxRetries is the default number of times to retry failed requests
	DefaultMaxRetries = 3
	// DefaultScrollDuration is the default duration for scroll operations
	DefaultScrollDuration = 5 * time.Minute
)

// Module provides the storage module for dependency injection.
var Module = fx.Module("storage",
	fx.Provide(
		// First provide the options
		fx.Annotate(
			func(cfg config.Interface) Options {
				return NewOptionsFromConfig(cfg)
			},
			fx.ParamTags(`name:"config"`),
		),
		// Then provide the Elasticsearch client
		fx.Annotate(
			func(cfg config.Interface, logger logger.Interface) (*es.Client, error) {
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
			},
			fx.ParamTags(`name:"config"`, ""),
		),
		// Then provide the index manager
		func(client *es.Client, logger logger.Interface) api.IndexManager {
			return NewElasticsearchIndexManager(client, logger)
		},
		// Finally provide the storage interface
		func(client *es.Client, logger logger.Interface, opts Options) storagetypes.Interface {
			return NewStorage(client, logger, opts)
		},
	),
)
