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
}

// DefaultOptions returns default options for ElasticsearchStorage
func DefaultOptions() Options {
	return Options{
		ScrollDuration: "5m",
		Transport:      http.DefaultTransport,
		IndexName:      "content", // Default index name
	}
}

// NewOptionsFromConfig creates Options from a config
func NewOptionsFromConfig(cfg config.Interface) Options {
	opts := DefaultOptions()
	esConfig := cfg.GetElasticsearchConfig()

	// Create transport with TLS config if needed
	if esConfig.TLS.SkipVerify {
		// Clone the default transport to preserve other settings
		if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport := defaultTransport.Clone()
			transport.TLSClientConfig = &tls.Config{
				//nolint:gosec // We are using the SkipVerify setting from the config
				InsecureSkipVerify: true,
			}
			opts.Transport = transport
		}
	}

	// Set values from config
	opts.Addresses = esConfig.Addresses
	opts.Username = esConfig.Username
	opts.Password = esConfig.Password
	opts.APIKey = esConfig.APIKey
	opts.IndexName = esConfig.IndexName

	return opts
}
