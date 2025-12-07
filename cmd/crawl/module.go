// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
)

// Common errors
var (
	ErrInvalidJob    = errors.New("invalid job")
	ErrInvalidJobURL = errors.New("invalid job URL")
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultInitTimeout is the default timeout for module initialization.
	DefaultInitTimeout = 30 * time.Second
)

// Module provides the crawl command module for dependency injection.
// Note: Command registration is handled by Command() function, not through FX Group annotation.
// The crawl command constructs dependencies directly, so no FX modules are needed here.
var Module = fx.Module("crawl")
