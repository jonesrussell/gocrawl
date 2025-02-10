package storage

import (
	"net/http"
)

// Options holds configuration options for ElasticsearchStorage
type Options struct {
	URL            string
	Username       string
	Password       string
	APIKey         string
	ScrollDuration string
	Transport      http.RoundTripper
}

// DefaultOptions returns default options for ElasticsearchStorage
func DefaultOptions() Options {
	return Options{
		ScrollDuration: "5m",
		Transport:      http.DefaultTransport,
	}
}
