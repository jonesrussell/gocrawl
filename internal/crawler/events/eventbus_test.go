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

// Debug logs a debug message.
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn logs a warning message.
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error logs an error message.
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal logs a fatal message and exits.
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With creates a new logger with the given fields.
func (m *MockLogger) With(fields ...any) logger.Interface {
	m.Called(fields)
	return m
}

// WithUser adds a user ID to the logger.
func (m *MockLogger) WithUser(userID string) logger.Interface {
	m.Called(userID)
	return m
}

// WithRequestID adds a request ID to the logger.
func (m *MockLogger) WithRequestID(requestID string) logger.Interface {
	m.Called(requestID)
	return m
}

// WithTraceID adds a trace ID to the logger.
func (m *MockLogger) WithTraceID(traceID string) logger.Interface {
	m.Called(traceID)
	return m
}

// WithSpanID adds a span ID to the logger.
func (m *MockLogger) WithSpanID(spanID string) logger.Interface {
	m.Called(spanID)
	return m
}

// WithDuration adds a duration to the logger.
func (m *MockLogger) WithDuration(duration time.Duration) logger.Interface {
	m.Called(duration)
	return m
}

// WithError adds an error to the logger.
func (m *MockLogger) WithError(err error) logger.Interface {
	m.Called(err)
	return m
}

// WithComponent adds a component name to the logger.
func (m *MockLogger) WithComponent(component string) logger.Interface {
	m.Called(component)
	return m
}

// WithVersion adds a version to the logger.
func (m *MockLogger) WithVersion(version string) logger.Interface {
	m.Called(version)
	return m
}

// WithEnvironment adds an environment to the logger.
func (m *MockLogger) WithEnvironment(env string) logger.Interface {
	m.Called(env)
	return m
}

// NewMockLogger creates a new mock logger instance
func NewMockLogger() *MockLogger {
	return &MockLogger{}
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

	log := NewMockLogger()
	bus := events.NewEventBus(log)

	t.Run("NewEventBus", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, bus)
	})

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		article := &models.Article{Title: "Test Article"}
		bus.PublishArticle(context.Background(), article)
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
		bus.PublishError(context.Background(), err)
		require.Eventually(t, func() bool {
			return handler.err != nil
		}, time.Second, time.Millisecond*100)
		assert.Equal(t, err, handler.err)
	})

	t.Run("PublishStart", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStart(context.Background())
		require.Eventually(t, func() bool {
			return handler.started
		}, time.Second, time.Millisecond*100)
	})

	t.Run("PublishStop", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStop(context.Background())
		require.Eventually(t, func() bool {
			return handler.stopped
		}, time.Second, time.Millisecond*100)
	})
}
