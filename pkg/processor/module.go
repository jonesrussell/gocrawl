// Package processor provides content processing functionality for the application.
package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/jonesrussell/gocrawl/pkg/logger"
)

// Module provides the processor module for dependency injection.
var Module = fx.Module("processor",
	fx.Provide(
		NewProcessor,
	),
)

// NewProcessor creates a new processor instance.
func NewProcessor(p Params) (Interface, error) {
	if err := ValidateParams(p); err != nil {
		return nil, err
	}
	return &processor{
		logger: p.Logger,
		metrics: &Metrics{
			LastProcessedTime: time.Now(),
		},
		htmlProcessor: NewHTMLProcessor(defaultSelectors),
	}, nil
}

var defaultSelectors = map[string]string{
	"title":        "h1",
	"body":         "article",
	"author":       ".author",
	"published_at": "time",
	"categories":   ".categories",
	"tags":         ".tags",
}

type processor struct {
	logger        logger.Interface
	metrics       *Metrics
	mu            sync.RWMutex
	htmlProcessor *HTMLProcessor
}

func (p *processor) Process(ctx context.Context, source string, content []byte) error {
	start := time.Now()
	defer func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.metrics.ProcessingDuration += time.Since(start)
	}()

	if err := p.Validate(source, content); err != nil {
		p.mu.Lock()
		p.metrics.ErrorCount++
		p.mu.Unlock()
		return err
	}

	// Process the content
	processed, err := p.htmlProcessor.Process(content)
	if err != nil {
		p.mu.Lock()
		p.metrics.ErrorCount++
		p.mu.Unlock()
		return fmt.Errorf("failed to process content: %w", err)
	}

	// Update the source
	processed.Content.Source = source

	// Log the processed content
	p.logger.Info("processed content",
		"source", source,
		"title", processed.Content.Title,
		"url", processed.Content.URL,
		"published_at", processed.Content.PublishedAt,
		"author", processed.Content.Author,
		"categories", processed.Content.Categories,
		"tags", processed.Content.Tags,
	)

	p.mu.Lock()
	p.metrics.ProcessedCount++
	p.metrics.LastProcessedTime = time.Now()
	p.mu.Unlock()

	return nil
}

func (p *processor) Validate(source string, content []byte) error {
	if source == "" {
		return fmt.Errorf("%w: source is required", ErrInvalidSource)
	}
	if len(content) == 0 {
		return fmt.Errorf("%w: content is required", ErrInvalidContent)
	}
	return nil
}

func (p *processor) GetMetrics() *Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.metrics
}
