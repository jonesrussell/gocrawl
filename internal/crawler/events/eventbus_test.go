package events

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockLogger implements logger.Interface for testing.
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger instance.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Info implements logger.Interface.
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements logger.Interface.
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Debug implements logger.Interface.
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements logger.Interface.
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements logger.Interface.
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With implements logger.Interface.
func (m *MockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	if result, ok := args.Get(0).(logger.Interface); ok {
		return result
	}
	return NewMockLogger()
}

// MockEventHandler is a mock implementation of EventHandler
type MockEventHandler struct {
	mock.Mock
}

func (m *MockEventHandler) HandleArticle(ctx context.Context, article *models.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m *MockEventHandler) HandleError(ctx context.Context, err error) error {
	args := m.Called(ctx, err)
	return args.Error(0)
}

func (m *MockEventHandler) HandleStart(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventHandler) HandleStop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestEventBus(t *testing.T) {
	t.Parallel()

	t.Run("NewEventBus", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		require.NotNil(t, bus)
	})

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		handler := &MockEventHandler{}
		bus.Subscribe(handler)
		assert.Len(t, bus.handlers, 1)
	})

	t.Run("PublishArticle", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		handler := &MockEventHandler{}

		article := &models.Article{
			ID:            "test-id",
			Title:         "Test Article",
			Body:          "Test Content",
			Source:        "http://test.com",
			PublishedDate: time.Now(),
		}

		handler.On("HandleArticle", mock.Anything, article).Return(nil)
		bus.Subscribe(handler)

		err := bus.PublishArticle(context.Background(), article)
		require.NoError(t, err)
		handler.AssertExpectations(t)
	})

	t.Run("PublishError", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		handler := &MockEventHandler{}

		testErr := assert.AnError
		handler.On("HandleError", mock.Anything, testErr).Return(nil)
		bus.Subscribe(handler)

		err := bus.PublishError(context.Background(), testErr)
		require.NoError(t, err)
		handler.AssertExpectations(t)
	})

	t.Run("PublishStart", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		handler := &MockEventHandler{}

		handler.On("HandleStart", mock.Anything).Return(nil)
		bus.Subscribe(handler)

		err := bus.PublishStart(context.Background())
		require.NoError(t, err)
		handler.AssertExpectations(t)
	})

	t.Run("PublishStop", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		handler := &MockEventHandler{}

		handler.On("HandleStop", mock.Anything).Return(nil)
		bus.Subscribe(handler)

		err := bus.PublishStop(context.Background())
		require.NoError(t, err)
		handler.AssertExpectations(t)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		ctx := context.Background()

		// Create multiple mock handlers
		handlers := make([]*MockEventHandler, 10)
		for i := range handlers {
			handlers[i] = &MockEventHandler{}
			handlers[i].On("HandleArticle", mock.Anything, mock.Anything).Return(nil)
			bus.Subscribe(handlers[i])
		}

		// Create a test article
		article := &models.Article{
			ID:            "test-id",
			Title:         "Test Article",
			Body:          "Test content",
			Source:        "https://example.com/test",
			PublishedDate: time.Now(),
		}

		// Start multiple goroutines to publish articles
		for i := 0; i < 100; i++ {
			go func() {
				bus.PublishArticle(ctx, article)
			}()
		}

		// Wait for all goroutines to complete
		time.Sleep(100 * time.Millisecond)

		// Verify all handlers were called
		for _, handler := range handlers {
			handler.AssertNumberOfCalls(t, "HandleArticle", 100)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		log := NewMockLogger()
		bus := NewEventBus(log)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Create a mock handler
		handler := &MockEventHandler{}
		handler.On("HandleArticle", mock.Anything, mock.Anything).Return(nil)

		// Subscribe handler
		bus.Subscribe(handler)

		// Create a test article
		article := &models.Article{
			ID:            "test-id",
			Title:         "Test Article",
			Body:          "Test content",
			Source:        "https://example.com/test",
			PublishedDate: time.Now(),
		}

		// Publish article with cancelled context
		bus.PublishArticle(ctx, article)

		// Verify handler was not called
		handler.AssertNotCalled(t, "HandleArticle", ctx, article)
	})
}
