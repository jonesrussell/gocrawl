// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
package sources

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

// Interface defines the interface for source management operations.
type Interface interface {
	// ListSources retrieves all sources.
	ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error)
	// AddSource adds a new source.
	AddSource(ctx context.Context, source *sourceutils.SourceConfig) error
	// UpdateSource updates an existing source.
	UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error
	// DeleteSource deletes a source by name.
	DeleteSource(ctx context.Context, name string) error
	// ValidateSource validates a source configuration.
	ValidateSource(source *sourceutils.SourceConfig) error
	// GetMetrics returns the current metrics.
	GetMetrics() sourceutils.SourcesMetrics
	// FindByName finds a source by name. Returns nil if not found.
	FindByName(name string) *sourceutils.SourceConfig
	// GetSources retrieves all source configurations.
	GetSources() ([]sourceutils.SourceConfig, error)
}

// Params contains the parameters for creating a new source manager.
type Params struct {
	// Logger is the logger to use.
	Logger logger.Interface
}

// ErrInvalidSource is returned when a source is invalid.
var ErrInvalidSource = errors.New("invalid source")

// ErrSourceNotFound is returned when a source is not found.
var ErrSourceNotFound = errors.New("source not found")

// ErrSourceExists is returned when a source already exists.
var ErrSourceExists = errors.New("source already exists")

// ValidateParams validates the parameters for creating a new source manager.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
