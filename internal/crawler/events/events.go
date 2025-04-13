// Package events provides event types and handlers for the crawler's event-based architecture.
package events

import (
	"context"
	"sync"
)

// ContentType represents the type of discovered content.
type ContentType string

const (
	// TypeArticle represents an article content type.
	TypeArticle ContentType = "article"
)

// Content represents discovered content from crawling.
type Content struct {
	// URL is the source URL of the content.
	URL string
	// Type is the content type (article, image, etc.).
	Type ContentType
	// Title is the content title if available.
	Title string
	// Description is a brief description if available.
	Description string
	// RawContent is the raw content data.
	RawContent string
	// Metadata contains additional content metadata.
	Metadata map[string]string
}

// Handler processes discovered content.
type Handler func(ctx context.Context, content *Content) error

// Bus manages content event subscriptions and publishing.
type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

// NewBus creates a new event bus instance.
func NewBus() *Bus {
	return &Bus{
		handlers: make([]Handler, 0),
	}
}

// Subscribe adds a new content handler.
func (b *Bus) Subscribe(handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// Publish sends content to all subscribed handlers.
func (b *Bus) Publish(ctx context.Context, content *Content) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var lastErr error
	for _, handler := range b.handlers {
		if err := handler(ctx, content); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
