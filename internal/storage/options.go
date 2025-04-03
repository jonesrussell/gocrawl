package storage

import (
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

// StorageOptions holds the configuration options for the storage implementation
type StorageOptions struct {
	IndexName string
	// Add other options as needed
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
func NewOptionsFromConfig(cfg config.Interface, transport http.RoundTripper) Options {
	opts := DefaultOptions()
	esConfig := cfg.GetElasticsearchConfig()

	// Use the provided transport
	opts.Transport = transport

	// Set values from config
	opts.Addresses = esConfig.Addresses
	opts.Username = esConfig.Username
	opts.Password = esConfig.Password
	opts.APIKey = esConfig.APIKey
	opts.IndexName = esConfig.IndexName

	return opts
}
