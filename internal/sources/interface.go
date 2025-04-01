// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
package sources

import (
	"context"
	"errors"
	"time"
)

// Interface defines the interface for source management.
type Interface interface {
	// GetSource retrieves a source by name.
	GetSource(ctx context.Context, name string) (*Config, error)
	// ListSources retrieves all sources.
	ListSources(ctx context.Context) ([]*Config, error)
	// AddSource adds a new source.
	AddSource(ctx context.Context, source *Config) error
	// UpdateSource updates an existing source.
	UpdateSource(ctx context.Context, source *Config) error
	// DeleteSource deletes a source by name.
	DeleteSource(ctx context.Context, name string) error
	// ValidateSource validates a source configuration.
	ValidateSource(source *Config) error
	// GetMetrics returns the current metrics.
	GetMetrics() Metrics
}

// Params defines the parameters for creating a new Sources instance.
type Params struct {
	// Logger is the logger to use.
	Logger interface {
		Debug(msg string, fields ...any)
		Info(msg string, fields ...any)
		Warn(msg string, fields ...any)
		Error(msg string, fields ...any)
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
	// ErrSourceNotFound is returned when a source is not found.
	ErrSourceNotFound = errors.New("source not found")
	// ErrInvalidSource is returned when a source configuration is invalid.
	ErrInvalidSource = errors.New("invalid source configuration")
)

// ValidateParams validates the parameters for creating a sources instance.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
