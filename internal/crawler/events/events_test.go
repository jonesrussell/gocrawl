// Package events_test provides tests for the events package.
package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBus_Subscribe tests the subscription functionality of the event bus.
func TestBus_Subscribe(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Test subscribing a handler
	handler := func(ctx context.Context, content *events.Content) error {
		return nil
	}
	bus.Subscribe(handler)

	// Verify handler was added
	content := &events.Content{
		URL:         "http://test.com",
		Type:        events.TypeArticle,
		Title:       "Test Article",
		Description: "Test Description",
		RawContent:  "Test Content",
		Metadata:    map[string]string{"key": "value"},
	}
	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)
}

// TestBus_Publish tests the publishing functionality of the event bus.
func TestBus_Publish(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Create a channel to receive published content
	received := make(chan *events.Content, 1)
	handler := func(ctx context.Context, content *events.Content) error {
		received <- content
		return nil
	}
	bus.Subscribe(handler)

	// Create test content
	content := &events.Content{
		URL:         "http://test.com",
		Type:        events.TypeArticle,
		Title:       "Test Article",
		Description: "Test Description",
		RawContent:  "Test Content",
		Metadata:    map[string]string{"key": "value"},
	}

	// Publish content
	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)

	// Wait for content to be received
	select {
	case receivedContent := <-received:
		assert.Equal(t, content.URL, receivedContent.URL)
		assert.Equal(t, content.Type, receivedContent.Type)
		assert.Equal(t, content.Title, receivedContent.Title)
		assert.Equal(t, content.Description, receivedContent.Description)
		assert.Equal(t, content.RawContent, receivedContent.RawContent)
		assert.Equal(t, content.Metadata, receivedContent.Metadata)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for content to be received")
	}
}

// TestBus_Publish_Error tests error handling in the event bus.
func TestBus_Publish_Error(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Create a handler that returns an error
	testErr := errors.New("test error")
	handler := func(ctx context.Context, content *events.Content) error {
		return testErr
	}
	bus.Subscribe(handler)

	// Create test content
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}

	// Publish content and verify error is returned
	err := bus.Publish(context.Background(), content)
	assert.Error(t, err)
	assert.Equal(t, testErr, err)
}

// TestBus_Publish_MultipleHandlers tests publishing to multiple handlers.
func TestBus_Publish_MultipleHandlers(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Create channels to receive published content
	received1 := make(chan *events.Content, 1)
	received2 := make(chan *events.Content, 1)

	// Create two handlers
	handler1 := func(ctx context.Context, content *events.Content) error {
		received1 <- content
		return nil
	}
	handler2 := func(ctx context.Context, content *events.Content) error {
		received2 <- content
		return nil
	}

	// Subscribe both handlers
	bus.Subscribe(handler1)
	bus.Subscribe(handler2)

	// Create test content
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}

	// Publish content
	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)

	// Wait for content to be received by both handlers
	select {
	case receivedContent1 := <-received1:
		assert.Equal(t, content.URL, receivedContent1.URL)
		assert.Equal(t, content.Type, receivedContent1.Type)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for content to be received by handler1")
	}

	select {
	case receivedContent2 := <-received2:
		assert.Equal(t, content.URL, receivedContent2.URL)
		assert.Equal(t, content.Type, receivedContent2.Type)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for content to be received by handler2")
	}
}

// TestBus_Publish_Concurrent tests concurrent publishing and subscription.
func TestBus_Publish_Concurrent(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Create a channel to receive published content
	received := make(chan *events.Content, 100)
	handler := func(ctx context.Context, content *events.Content) error {
		received <- content
		return nil
	}
	bus.Subscribe(handler)

	// Create test content
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}

	// Publish content concurrently
	for i := 0; i < 100; i++ {
		go func() {
			err := bus.Publish(context.Background(), content)
			require.NoError(t, err)
		}()
	}

	// Wait for all content to be received
	for i := 0; i < 100; i++ {
		select {
		case receivedContent := <-received:
			assert.Equal(t, content.URL, receivedContent.URL)
			assert.Equal(t, content.Type, receivedContent.Type)
		case <-time.After(time.Second):
			t.Fatalf("Timeout waiting for content %d to be received", i)
		}
	}
}

// TestBus_Publish_ContextCancellation tests context cancellation during publishing.
func TestBus_Publish_ContextCancellation(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	require.NotNil(t, bus)

	// Create a handler that blocks
	handler := func(ctx context.Context, content *events.Content) error {
		<-ctx.Done()
		return ctx.Err()
	}
	bus.Subscribe(handler)

	// Create test content
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Publish content with cancelled context
	err := bus.Publish(ctx, content)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
