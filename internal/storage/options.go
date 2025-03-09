package storage

import (
	"crypto/tls"
	"net/http"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// Options holds configuration options for ElasticsearchStorage
type Options struct {
	Addresses      []string
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
func NewOptionsFromConfig(cfg config.Interface) Options {
	opts := DefaultOptions()
	esConfig := cfg.GetElasticsearchConfig()

	// Create transport with TLS config if needed
	if esConfig.TLS.SkipVerify {
		opts.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec // We are using the SkipVerify setting from the config
				InsecureSkipVerify: true,
			},
		}
	}

	// Set values from config
	opts.Addresses = esConfig.Addresses
	opts.Username = esConfig.Username
	opts.Password = esConfig.Password
	opts.APIKey = esConfig.APIKey
	opts.IndexName = esConfig.IndexName
	opts.SkipTLS = esConfig.TLS.SkipVerify

	return opts
}
