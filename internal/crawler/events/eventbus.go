package events

import (
	"context"
	"sync"

	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// EventBus implements the crawler.EventBus interface for managing event distribution.
type EventBus struct {
	mu       sync.RWMutex
	handlers []EventHandler
	logger   logger.Interface
}

// NewEventBus creates a new EventBus instance.
func NewEventBus(log logger.Interface) *EventBus {
	return &EventBus{
		handlers: make([]EventHandler, 0),
		logger:   log,
	}
}

// Subscribe adds an event handler to the bus.
func (b *EventBus) Subscribe(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// PublishArticle publishes an article event to all handlers.
// Thread-safe: uses read lock and copies handlers slice.
func (b *EventBus) PublishArticle(ctx context.Context, article *domain.Article) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Get a snapshot of handlers under read lock
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	// Dispatch to handlers without holding lock
	for _, handler := range handlers {
		if err := handler.HandleArticle(ctx, article); err != nil {
			b.logger.Error("failed to handle article event",
				"error", err,
				"articleID", article.ID,
				"url", article.Source,
			)
		}
	}
	return nil
}

// PublishError publishes an error event to all handlers.
// Thread-safe: uses read lock and copies handlers slice.
func (b *EventBus) PublishError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return
	}

	// Get a snapshot of handlers under read lock
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	// Dispatch to handlers without holding lock
	for _, handler := range handlers {
		handlerErr := handler.HandleError(ctx, err)
		if handlerErr != nil {
			b.logger.Error("Failed to handle error",
				"error", handlerErr,
				"original_error", err,
			)
		}
	}
}

// PublishStart publishes a start event to all handlers.
// Thread-safe: uses read lock and copies handlers slice.
func (b *EventBus) PublishStart(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Get a snapshot of handlers under read lock
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	// Dispatch to handlers without holding lock
	for _, handler := range handlers {
		if err := handler.HandleStart(ctx); err != nil {
			b.logger.Error("failed to handle start event",
				"error", err,
			)
		}
	}
	return nil
}

// PublishStop publishes a stop event to all handlers.
// Thread-safe: uses read lock and copies handlers slice.
func (b *EventBus) PublishStop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Get a snapshot of handlers under read lock
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	// Dispatch to handlers without holding lock
	for _, handler := range handlers {
		if err := handler.HandleStop(ctx); err != nil {
			b.logger.Error("failed to handle stop event",
				"error", err,
			)
		}
	}
	return nil
}

// HandleError handles an error that occurred during event processing.
func (b *EventBus) HandleError(err error, article *domain.Article) {
	b.logger.Error("Error processing article",
		"error", err,
		"articleID", article.ID,
		"url", article.Source,
	)
}

// HandleHandlerError handles an error that occurred in an event handler.
func (b *EventBus) HandleHandlerError(handlerErr, err error) {
	b.logger.Error("Error in event handler",
		"error", handlerErr,
		"original_error", err,
	)
}

// HandlerCount returns the number of registered handlers.
// Thread-safe: uses read lock.
func (b *EventBus) HandlerCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers)
}
