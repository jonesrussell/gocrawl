// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
package sources

import (
	"context"
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common/types"
)

// Interface defines the interface for source management.
type Interface interface {
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
	// FindByName finds a source by name.
	FindByName(name string) (*Config, error)
	// GetSources retrieves all source configurations.
	GetSources() ([]Config, error)
}

// Params defines the parameters for creating a new Sources instance.
type Params struct {
	// Logger is the logger to use.
	Logger types.Logger
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

// ValidateParams validates the parameters for creating a new Sources instance.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
