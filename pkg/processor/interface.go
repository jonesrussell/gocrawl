// Package processor provides content processing functionality for the application.
package processor

import (
	"context"
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
)

var (
	// ErrInvalidSource is returned when the source is invalid.
	ErrInvalidSource = errors.New("invalid source")
	// ErrInvalidContent is returned when the content is invalid.
	ErrInvalidContent = errors.New("invalid content")
)

// Interface defines the interface for content processing operations.
// It provides methods for processing and managing content from various sources.
type Interface interface {
	// Process processes content from a source.
	Process(ctx context.Context, source string, content []byte) error
	// Validate validates a source's content.
	Validate(source string, content []byte) error
	// GetMetrics returns the current processing metrics.
	GetMetrics() *Metrics
}

// Metrics holds the processing metrics.
type Metrics struct {
	// ProcessedCount is the number of items processed.
	ProcessedCount int64
	// ErrorCount is the number of processing errors.
	ErrorCount int64
	// LastProcessedTime is the time of the last successful processing.
	LastProcessedTime time.Time
	// ProcessingDuration is the total time spent processing.
	ProcessingDuration time.Duration
}

// Params holds the parameters for creating a processor.
type Params struct {
	Logger common.Logger
}

// ValidateParams validates the parameters for creating a processor instance.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
