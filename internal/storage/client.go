// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

// ClientModule provides the Elasticsearch client
var ClientModule = fx.Module("elasticsearch",
	fx.Provide(
		func(cfg config.Interface) (*es.Client, error) {
			esConfig := cfg.GetElasticsearchConfig()
			if esConfig == nil {
				return nil, errors.New("elasticsearch configuration is required")
			}

			transport := CreateTransport(esConfig)
			clientConfig := CreateClientConfig(esConfig, transport)

			client, err := es.NewClient(clientConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
			}

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
	),
)
