package storage

import (
	"crypto/tls"
	"net/http"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// Options holds configuration options for ElasticsearchStorage
type Options struct {
	URL            string
	Username       string
	Password       string
	APIKey         string
	ScrollDuration string
	Transport      http.RoundTripper
	IndexName      string // Name of the index to use for content
	SkipTLS        bool   // Whether to skip TLS verification
}

// DefaultOptions returns default options for ElasticsearchStorage
func DefaultOptions() Options {
	return Options{
		ScrollDuration: "5m",
		Transport:      http.DefaultTransport,
		IndexName:      "content", // Default index name
		SkipTLS:        false,     // Default to secure TLS
	}
}

// NewOptionsFromConfig creates Options from a config
func NewOptionsFromConfig(cfg *config.Config) Options {
	opts := DefaultOptions()

	// Create transport with TLS config if needed
	if cfg.Elasticsearch.SkipTLS {
		opts.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec // We are using the SkipTLS setting from the config
				InsecureSkipVerify: true,
			},
		}
	}

	// Set values from config
	opts.URL = cfg.Elasticsearch.URL
	opts.Username = cfg.Elasticsearch.Username
	opts.Password = cfg.Elasticsearch.Password
	opts.APIKey = cfg.Elasticsearch.APIKey
	opts.IndexName = cfg.Elasticsearch.IndexName
	opts.SkipTLS = cfg.Elasticsearch.SkipTLS

	return opts
}
