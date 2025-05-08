package storage

import (
	"context"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ElasticsearchStorage implements the storage interface using Elasticsearch
type ElasticsearchStorage struct {
	client *es.Client
	config *config.Config
	logger logger.Interface
}

// NewElasticsearchStorage creates a new Elasticsearch storage instance
func NewElasticsearchStorage(client *es.Client, config *config.Config, logger logger.Interface) *ElasticsearchStorage {
	return &ElasticsearchStorage{
		client: client,
		config: config,
		logger: logger,
	}
}

// TestConnection tests the connection to Elasticsearch.
func (s *ElasticsearchStorage) TestConnection(ctx context.Context) error {
	s.logger.Debug("Testing Elasticsearch connection",
		"addresses", s.config.Elasticsearch.Addresses,
		"username", s.config.Elasticsearch.Username,
		"api_key", s.config.Elasticsearch.APIKey != "",
		"tls_insecure_skip_verify", s.config.Elasticsearch.TLS != nil && s.config.Elasticsearch.TLS.InsecureSkipVerify,
		"tls_has_ca_file", s.config.Elasticsearch.TLS != nil && s.config.Elasticsearch.TLS.CAFile != "",
		"tls_has_cert_file", s.config.Elasticsearch.TLS != nil && s.config.Elasticsearch.TLS.CertFile != "",
		"tls_has_key_file", s.config.Elasticsearch.TLS != nil && s.config.Elasticsearch.TLS.KeyFile != "")

	res, err := s.client.Info()
	if err != nil {
		s.logger.Error("Failed to connect to Elasticsearch", "error", err)
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Error response from Elasticsearch", "response", res.String())
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	s.logger.Info("Successfully connected to Elasticsearch")
	return nil
}

// Client implements the storage interface for Elasticsearch
type Client struct {
	client *es.Client
	opts   Options
}

// NewClient creates a new Elasticsearch client
func NewClient(opts Options) (*Client, error) {
	cfg := es.Config{
		Addresses: opts.Addresses,
		Username:  opts.Username,
		Password:  opts.Password,
		APIKey:    opts.APIKey,
		Transport: opts.Transport,
	}

	client, err := es.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %w", err)
	}

	return &Client{
		client: client,
		opts:   opts,
	}, nil
}

// GetClient returns the Elasticsearch client
func (c *Client) GetClient() *es.Client {
	return c.client
}
