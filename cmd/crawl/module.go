// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
	// DefaultZapFieldsCapacity is the default capacity for zap fields slice
	DefaultZapFieldsCapacity = 2
)

// Module provides the crawl command module for dependency injection
var Module = fx.Options(
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	sources.Module,
	crawler.Module,

	// Provide the context
	fx.Provide(context.Background),

	// Provide the done channel
	fx.Provide(func() chan struct{} {
		return make(chan struct{})
	}),

	// Provide the job service
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			storage types.Interface,
			sources sources.Interface,
			crawler crawler.Interface,
			done chan struct{},
			processorFactory crawler.ProcessorFactory,
		) common.JobService {
			return NewJobService(JobServiceParams{
				Logger:           logger,
				Sources:          sources,
				Crawler:          crawler,
				Done:             done,
				Storage:          storage,
				ProcessorFactory: processorFactory,
			})
		},
		fx.As(new(common.JobService)),
	)),

	// Provide the job processor
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			config config.Interface,
		) *DefaultJobService {
			return &DefaultJobService{
				logger: logger,
				config: config,
			}
		},
		fx.As(new(content.ContentProcessor)),
	)),
)

// NewCommand creates a new crawl command.
func NewCommand(p struct {
	fx.In
	Crawler crawler.Interface
}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a website",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Crawler.Start(context.Background(), "default")
		},
	}
	return cmd
}

// DefaultJobService implements the content.ContentProcessor interface.
type DefaultJobService struct {
	logger logger.Interface
	config config.Interface
}

// ValidateJob validates a job before processing.
func (s *DefaultJobService) ValidateJob(job *content.Job) error {
	if job == nil {
		return ErrInvalidJob
	}
	if job.URL == "" {
		return ErrInvalidJobURL
	}
	return nil
}

// Process implements content.ContentProcessor.
func (s *DefaultJobService) Process(ctx context.Context, content any) error {
	if content == nil {
		return ErrInvalidJob
	}
	// TODO: Implement job processing
	return nil
}

// CanProcess implements content.ContentProcessor.
func (s *DefaultJobService) CanProcess(contentType contenttype.Type) bool {
	return contentType == contenttype.Job
}

// ContentType implements content.ContentProcessor.
func (s *DefaultJobService) ContentType() contenttype.Type {
	return contenttype.Job
}

// ExtractContent implements content.ContentProcessor.
func (s *DefaultJobService) ExtractContent() (string, error) {
	return "", errors.New("not implemented")
}

// ExtractLinks implements content.ContentProcessor.
func (s *DefaultJobService) ExtractLinks() ([]string, error) {
	return nil, errors.New("not implemented")
}

// GetProcessor implements content.ContentProcessor.
func (s *DefaultJobService) GetProcessor(contentType contenttype.Type) (content.ContentProcessor, error) {
	if contentType != contenttype.Job {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return s, nil
}

// ParseHTML implements content.ContentProcessor.
func (s *DefaultJobService) ParseHTML(r io.Reader) error {
	return errors.New("not implemented")
}

// RegisterProcessor implements content.ContentProcessor.
func (s *DefaultJobService) RegisterProcessor(processor content.Processor) {
	// No-op for now
}

// Start implements content.ContentProcessor.
func (s *DefaultJobService) Start(ctx context.Context) error {
	return nil
}

// Stop implements content.ContentProcessor.
func (s *DefaultJobService) Stop(ctx context.Context) error {
	return nil
}

// ZapWrapper wraps a zap.Logger to implement logger.Interface.
type ZapWrapper struct {
	*zap.Logger
}

// Debug implements logger.Interface.
func (l *ZapWrapper) Debug(msg string, fields ...any) {
	l.Logger.Debug(msg, toZapFields(fields)...)
}

// Info implements logger.Interface.
func (l *ZapWrapper) Info(msg string, fields ...any) {
	l.Logger.Info(msg, toZapFields(fields)...)
}

// Error implements logger.Interface.
func (l *ZapWrapper) Error(msg string, fields ...any) {
	l.Logger.Error(msg, toZapFields(fields)...)
}

// Warn implements logger.Interface.
func (l *ZapWrapper) Warn(msg string, fields ...any) {
	l.Logger.Warn(msg, toZapFields(fields)...)
}

// Fatal implements logger.Interface.
func (l *ZapWrapper) Fatal(msg string, fields ...any) {
	l.Logger.Fatal(msg, toZapFields(fields)...)
}

// toZapFields converts the fields to zap fields.
func toZapFields(fields []any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields)/DefaultZapFieldsCapacity)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return zapFields
}
