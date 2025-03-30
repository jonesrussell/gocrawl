// Package sources provides source management functionality for the application.
package sources

import (
	"context"
	"errors"
	"time"
)

// Source represents a content source configuration.
type Source struct {
	// Name is the unique identifier for the source.
	Name string `json:"name" yaml:"name"`
	// URL is the base URL to crawl.
	URL string `json:"url" yaml:"url"`
	// MaxDepth is the maximum depth to crawl.
	MaxDepth int `json:"max_depth" yaml:"max_depth"`
	// Time defines time-related fields to extract.
	Time struct {
		// PublishedAt is the selector for the published date.
		PublishedAt string `json:"published_at" yaml:"published_at"`
		// UpdatedAt is the selector for the last updated date.
		UpdatedAt string `json:"updated_at" yaml:"updated_at"`
	} `json:"time" yaml:"time"`
}

// Interface defines the interface for source management.
type Interface interface {
	// GetSource retrieves a source by name.
	GetSource(ctx context.Context, name string) (*Source, error)
	// ListSources retrieves all sources.
	ListSources(ctx context.Context) ([]*Source, error)
	// AddSource adds a new source.
	AddSource(ctx context.Context, source *Source) error
	// UpdateSource updates an existing source.
	UpdateSource(ctx context.Context, source *Source) error
	// DeleteSource deletes a source by name.
	DeleteSource(ctx context.Context, name string) error
	// ValidateSource validates a source configuration.
	ValidateSource(source *Source) error
	// GetMetrics returns the current metrics.
	GetMetrics() Metrics
}

// Params defines the parameters for creating a new Sources instance.
type Params struct {
	// Logger is the logger to use.
	Logger interface {
		Debug(msg string, fields ...interface{})
		Info(msg string, fields ...interface{})
		Warn(msg string, fields ...interface{})
		Error(msg string, fields ...interface{})
	}
}

// Metrics defines the metrics for the Sources module.
type Metrics struct {
	// SourceCount is the number of sources.
	SourceCount int64
	// LastUpdated is the last time a source was updated.
	LastUpdated time.Time
}

var (
	// ErrSourceExists is returned when a source already exists.
	ErrSourceExists = errors.New("source already exists")
)

// Config represents a source configuration.
type Config struct {
	// Name is the unique identifier for the source.
	Name string
	// URL is the base URL for the source.
	URL string
	// RateLimit is the rate limit for requests to this source.
	RateLimit time.Duration
	// MaxDepth is the maximum depth for crawling this source.
	MaxDepth int
	// Time is the list of times when this source should be crawled.
	Time []string
}

// ValidateParams validates the parameters for creating a sources instance.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
