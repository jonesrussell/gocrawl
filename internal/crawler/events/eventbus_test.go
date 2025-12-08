package events_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/domain"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

// MockEventHandler is a mock implementation of EventHandler
// Thread-safe for concurrent access.
type MockEventHandler struct {
	mu      sync.RWMutex
	article *domain.Article
	err     error
	started bool
	stopped bool
}

func (h *MockEventHandler) HandleArticle(ctx context.Context, article *domain.Article) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.article = article
	return nil
}

func (h *MockEventHandler) HandleError(ctx context.Context, err error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.err = err
	return nil
}

func (h *MockEventHandler) HandleStart(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.started = true
	return nil
}

func (h *MockEventHandler) HandleStop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopped = true
	return nil
}

// GetError returns the last error handled (thread-safe).
func (h *MockEventHandler) GetError() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.err
}

// GetArticle returns the last article handled (thread-safe).
func (h *MockEventHandler) GetArticle() *domain.Article {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.article
}

// IsStarted returns whether the start event was handled (thread-safe).
func (h *MockEventHandler) IsStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.started
}

// IsStopped returns whether the stop event was handled (thread-safe).
func (h *MockEventHandler) IsStopped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stopped
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
		article := &domain.Article{Title: "Test Article"}
		bus.PublishArticle(context.Background(), article)
		require.Eventually(t, func() bool {
			return handler.GetArticle() != nil
		}, time.Second, time.Millisecond*100)
		assert.Equal(t, article, handler.GetArticle())
	})

	t.Run("PublishError", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		err := errors.New("test error")
		bus.PublishError(context.Background(), err)
		require.Eventually(t, func() bool {
			return handler.GetError() != nil
		}, time.Second, time.Millisecond*100)
		assert.Equal(t, err, handler.GetError())
	})

	t.Run("PublishStart", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStart(context.Background())
		require.Eventually(t, func() bool {
			return handler.IsStarted()
		}, time.Second, time.Millisecond*100)
	})

	t.Run("PublishStop", func(t *testing.T) {
		t.Parallel()
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		bus.PublishStop(context.Background())
		require.Eventually(t, func() bool {
			return handler.IsStopped()
		}, time.Second, time.Millisecond*100)
	})
}

// TestEventBus_ConcurrentSubscribe tests concurrent Subscribe calls
func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	t.Parallel()

	log := NewMockLogger()
	bus := events.NewEventBus(log)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Subscribe(&MockEventHandler{})
		}()
	}

	wg.Wait()
	assert.Equal(t, 100, bus.HandlerCount())
}

// TestEventBus_ConcurrentPublish tests concurrent Publish calls
func TestEventBus_ConcurrentPublish(t *testing.T) {
	t.Parallel()

	log := NewMockLogger()
	bus := events.NewEventBus(log)
	handler := &MockEventHandler{}
	bus.Subscribe(handler)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.PublishError(context.Background(), errors.New("test"))
		}()
	}

	wg.Wait()
	// Should not panic or race
}

// TestEventBus_ConcurrentSubscribeAndPublish tests Subscribe and Publish concurrently
func TestEventBus_ConcurrentSubscribeAndPublish(t *testing.T) {
	t.Parallel()

	log := NewMockLogger()
	bus := events.NewEventBus(log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Publishers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					bus.PublishError(context.Background(), errors.New("test"))
				}
			}
		}()
	}

	// Subscribers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					bus.Subscribe(&MockEventHandler{})
				}
			}
		}()
	}

	wg.Wait()
}
