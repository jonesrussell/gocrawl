package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxevent"
)

// MockLogger is a mock implementation of logger.Interface
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger instance
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Debug implements logger.Interface
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Info implements logger.Interface
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements logger.Interface
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements logger.Interface
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements logger.Interface
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With implements logger.Interface
func (m *MockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	if result, ok := args.Get(0).(logger.Interface); ok {
		return result
	}
	return m
}

// NewFxLogger implements logger.Interface
func (m *MockLogger) NewFxLogger() fxevent.Logger {
	args := m.Called()
	if result, ok := args.Get(0).(fxevent.Logger); ok {
		return result
	}
	return &fxevent.NopLogger
}

// MockEventHandler is a mock implementation of EventHandler
type MockEventHandler struct {
	article *models.Article
	err     error
	started bool
	stopped bool
}

func (h *MockEventHandler) HandleArticle(ctx context.Context, article *models.Article) error {
	h.article = article
	return nil
}

func (h *MockEventHandler) HandleError(ctx context.Context, err error) error {
	h.err = err
	return nil
}

func (h *MockEventHandler) HandleStart(ctx context.Context) error {
	h.started = true
	return nil
}

func (h *MockEventHandler) HandleStop(ctx context.Context) error {
	h.stopped = true
	return nil
}

func TestEventBus(t *testing.T) {
	t.Parallel()

	logger := NewMockLogger()
	bus := events.NewEventBus(logger)

	t.Run("NewEventBus", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, bus)
	})

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		article := &models.Article{Title: "Test Article"}
		bus.PublishArticle(t.Context(), article)
		require.Eventually(t, func() bool {
			return handler.article != nil
		}, time.Second, time.Millisecond*100)
		assert.Equal(t, article, handler.article)
	})

	t.Run("PublishError", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		err := errors.New("test error")
		bus.PublishError(t.Context(), err)
		require.Eventually(t, func() bool {
			return handler.err != nil
		}, time.Second, time.Millisecond*100)
		assert.Equal(t, err, handler.err)
	})

	t.Run("PublishStart", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStart(t.Context())
		require.Eventually(t, func() bool {
			return handler.started
		}, time.Second, time.Millisecond*100)
	})

	t.Run("PublishStop", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStop(t.Context())
		require.Eventually(t, func() bool {
			return handler.stopped
		}, time.Second, time.Millisecond*100)
	})
}
