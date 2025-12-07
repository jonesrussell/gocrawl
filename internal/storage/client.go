// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// ClientParams contains dependencies for creating the Elasticsearch client
type ClientParams struct {
	fx.In

	Config config.Interface
	Logger logger.Interface
}

// ClientResult contains the Elasticsearch client
type ClientResult struct {
	fx.Out

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
	client, err := es.NewClient(clientConfig)
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

// ClientModule provides the Elasticsearch client
var ClientModule = fx.Module("elasticsearch-client",
	fx.Provide(
		NewClient,
	),
)
