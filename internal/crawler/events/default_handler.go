package events

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// DefaultHandler provides a basic implementation of EventHandler that logs events.
type DefaultHandler struct {
	logger logger.Interface
}

// NewDefaultHandler creates a new DefaultHandler instance.
func NewDefaultHandler(log logger.Interface) EventHandler {
	return &DefaultHandler{
		logger: log,
	}
}

// HandleArticle logs article events.
func (h *DefaultHandler) HandleArticle(ctx context.Context, article *domain.Article) error {
	h.logger.Info("Article processed",
		"id", article.ID,
		"title", article.Title,
		"url", article.Source,
		"component", "crawler",
	)
	return nil
}

// HandleError logs error events.
func (h *DefaultHandler) HandleError(ctx context.Context, err error) error {
	h.logger.Error("Error occurred",
		"error", err,
		"component", "crawler",
	)
	return nil
}

// HandleStart logs start events.
func (h *DefaultHandler) HandleStart(ctx context.Context) error {
	h.logger.Info("Crawler started",
		"component", "crawler",
	)
	return nil
}

// HandleStop logs stop events.
func (h *DefaultHandler) HandleStop(ctx context.Context) error {
	h.logger.Info("Crawler stopped",
		"component", "crawler",
	)
	return nil
}
