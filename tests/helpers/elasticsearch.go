// Package helpers provides testing utilities for integration tests.
package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultElasticsearchStartupTimeout is the default timeout for Elasticsearch to start.
	DefaultElasticsearchStartupTimeout = 60 * time.Second
	// DefaultHTTPClientTimeout is the default timeout for HTTP client requests.
	DefaultHTTPClientTimeout = 5 * time.Second
	// DefaultMaxRetries is the default number of retries for Elasticsearch health checks.
	DefaultMaxRetries = 30
	// HTTPStatusOK is the HTTP status code for successful requests.
	HTTPStatusOK = 200
)

// ElasticsearchContainer manages a test Elasticsearch instance.
type ElasticsearchContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	Address   string
}

// StartElasticsearch starts an Elasticsearch container for testing.
// It returns a container instance that should be stopped with Stop().
func StartElasticsearch(ctx context.Context) (*ElasticsearchContainer, error) {
	// Create Elasticsearch container with default configuration
	esContainer, err := elasticsearch.Run(
		ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.11.0",
		elasticsearch.WithPassword("changeme"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").WithPort("9200/tcp").WithStartupTimeout(DefaultElasticsearchStartupTimeout),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start Elasticsearch container: %w", err)
	}

	// Get container host and port
	host, err := esContainer.Host(ctx)
	if err != nil {
		_ = esContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := esContainer.MappedPort(ctx, "9200")
	if err != nil {
		_ = esContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Use net.JoinHostPort for proper host:port construction
	hostPort := net.JoinHostPort(host, mappedPort.Port())
	address := fmt.Sprintf("http://%s", hostPort)

	// Wait for Elasticsearch to be ready
	if waitErr := waitForElasticsearch(ctx, address); waitErr != nil {
		_ = esContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to wait for Elasticsearch: %w", waitErr)
	}

	return &ElasticsearchContainer{
		Container: esContainer,
		Host:      host,
		Port:      mappedPort.Port(),
		Address:   address,
	}, nil
}

// Stop stops and removes the Elasticsearch container.
func (e *ElasticsearchContainer) Stop(ctx context.Context) error {
	if e.Container == nil {
		return nil
	}
	return e.Container.Terminate(ctx)
}

// GetAddresses returns the Elasticsearch addresses as a slice.
// This matches the format expected by the Elasticsearch config.
func (e *ElasticsearchContainer) GetAddresses() []string {
	return []string{e.Address}
}

// waitForElasticsearch waits for Elasticsearch to be ready by pinging it.
func waitForElasticsearch(ctx context.Context, address string) error {
	client := &http.Client{
		Timeout: DefaultHTTPClientTimeout,
	}

	// Create request with basic auth for Elasticsearch 8.x
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/_cluster/health", address), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth("elastic", "changeme")

	for i := range DefaultMaxRetries {
		_ = i // Index not used, but required for range syntax
		resp, doErr := client.Do(req)
		if doErr == nil {
			resp.Body.Close()
			if resp.StatusCode == HTTPStatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			// Continue retrying
		}
	}

	return fmt.Errorf("elasticsearch did not become ready within %d seconds", DefaultMaxRetries)
}
