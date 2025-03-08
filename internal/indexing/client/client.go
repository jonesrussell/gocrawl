// Package client provides an Elasticsearch client wrapper.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
)

// Config holds the Elasticsearch client configuration.
type Config struct {
	Addresses []string
	Username  string
	Password  string
	TLSConfig *tls.Config
	Transport http.RoundTripper
	Retry     struct {
		MaxWait    time.Duration
		MaxRetries int
	}
}

// Client wraps the Elasticsearch client with additional functionality.
type Client struct {
	client *elasticsearch.Client
	logger common.Logger
	config *Config
}

// New creates a new Elasticsearch client with the given configuration.
func New(cfg *config.Config, logger common.Logger) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Elasticsearch.Addresses,
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: es,
		logger: logger,
		config: &Config{
			Addresses: cfg.Elasticsearch.Addresses,
			Username:  cfg.Elasticsearch.Username,
			Password:  cfg.Elasticsearch.Password,
			Retry: struct {
				MaxWait    time.Duration
				MaxRetries int
			}{
				MaxWait:    time.Second * 30,
				MaxRetries: 3,
			},
		},
	}, nil
}

// Ping checks if the Elasticsearch cluster is available.
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.client.Ping(c.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping failed: %s", res.String())
	}

	return nil
}

// Client returns the underlying Elasticsearch client.
func (c *Client) Client() *elasticsearch.Client {
	return c.client
}

// Config returns the client configuration.
func (c *Client) Config() *Config {
	return c.config
}

// Error represents an Elasticsearch error response.
type Error struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

// NewClient creates a new Elasticsearch client.
func NewClient(cfg *config.Config, logger common.Logger) (*Client, error) {
	return New(cfg, logger)
}

// Logger returns the logger instance.
func (c *Client) Logger() common.Logger {
	return c.logger
}
