package events

import (
	"context"
	"sync"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// EventBus implements the crawler.EventBus interface for managing event distribution.
type EventBus struct {
	mu       sync.RWMutex
	handlers []EventHandler
	logger   logger.Interface
}

// NewEventBus creates a new EventBus instance.
func NewEventBus(logger logger.Interface) *EventBus {
	return &EventBus{
		handlers: make([]EventHandler, 0),
		logger:   logger,
	}
}

// Subscribe adds an event handler to the bus.
func (b *EventBus) Subscribe(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// PublishArticle publishes an article event to all handlers.
func (b *EventBus) PublishArticle(ctx context.Context, article *models.Article) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers {
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
func (b *EventBus) PublishError(ctx context.Context, err error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers {
		if err := handler.HandleError(ctx, err); err != nil {
			b.logger.Error("failed to handle error event",
				"error", err,
			)
		}
	}
	return nil
}

// PublishStart publishes a start event to all handlers.
func (b *EventBus) PublishStart(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers {
		if err := handler.HandleStart(ctx); err != nil {
			b.logger.Error("failed to handle start event",
				"error", err,
			)
		}
	}
	return nil
}

// PublishStop publishes a stop event to all handlers.
func (b *EventBus) PublishStop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, handler := range b.handlers {
		if err := handler.HandleStop(ctx); err != nil {
			b.logger.Error("failed to handle stop event",
				"error", err,
			)
		}
	}
	return nil
}
