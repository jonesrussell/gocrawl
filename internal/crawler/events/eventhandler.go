package events

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/domain"
)

// EventHandler defines the interface for handling events from the EventBus.
type EventHandler interface {
	// HandleArticle processes an article event.
	HandleArticle(ctx context.Context, article *domain.Article) error

	// HandleError processes an error event.
	HandleError(ctx context.Context, err error) error

	// HandleStart processes a start event.
	HandleStart(ctx context.Context) error

	// HandleStop processes a stop event.
	HandleStop(ctx context.Context) error
}
