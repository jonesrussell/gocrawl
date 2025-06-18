// Package events_test implements tests for the events package.
package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/stretchr/testify/require"
)

// TestBus_Subscribe tests the subscription functionality of the event bus.
func TestBus_Subscribe(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	content := &events.Content{
		URL:         "http://test.com",
		Type:        events.TypeArticle,
		Title:       "Test Article",
		Description: "Test Description",
		RawContent:  "Test Content",
		Metadata:    map[string]string{"key": "value"},
	}
	received := make(chan *events.Content, 1)

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		received <- c
		return nil
	})

	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)

	select {
	case c := <-received:
		require.Equal(t, content.URL, c.URL)
		require.Equal(t, content.Type, c.Type)
		require.Equal(t, content.Title, c.Title)
		require.Equal(t, content.Description, c.Description)
		require.Equal(t, content.RawContent, c.RawContent)
		require.Equal(t, content.Metadata, c.Metadata)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for content")
	}
}

// TestBus_Publish tests the publishing functionality of the event bus.
func TestBus_Publish(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	content := &events.Content{
		URL:         "http://test.com",
		Type:        events.TypeArticle,
		Title:       "Test Article",
		Description: "Test Description",
		RawContent:  "Test Content",
		Metadata:    map[string]string{"key": "value"},
	}
	received := make(chan *events.Content, 1)

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		received <- c
		return nil
	})

	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)

	select {
	case c := <-received:
		require.Equal(t, content.URL, c.URL)
		require.Equal(t, content.Type, c.Type)
		require.Equal(t, content.Title, c.Title)
		require.Equal(t, content.Description, c.Description)
		require.Equal(t, content.RawContent, c.RawContent)
		require.Equal(t, content.Metadata, c.Metadata)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for content")
	}
}

// TestBus_Publish_Error tests error handling when a handler returns an error.
func TestBus_Publish_Error(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}
	testErr := errors.New("test error")

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		return testErr
	})

	err := bus.Publish(context.Background(), content)
	require.Error(t, err)
	require.Equal(t, testErr, err)
}

// TestBus_Publish_MultipleHandlers tests that multiple subscribers receive the published content.
func TestBus_Publish_MultipleHandlers(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}
	received1 := make(chan *events.Content, 1)
	received2 := make(chan *events.Content, 1)

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		received1 <- c
		return nil
	})

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		received2 <- c
		return nil
	})

	err := bus.Publish(context.Background(), content)
	require.NoError(t, err)

	select {
	case c := <-received1:
		require.Equal(t, content.URL, c.URL)
		require.Equal(t, content.Type, c.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for content in first handler")
	}

	select {
	case c := <-received2:
		require.Equal(t, content.URL, c.URL)
		require.Equal(t, content.Type, c.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for content in second handler")
	}
}

// TestBus_Publish_Concurrent tests concurrent publishing to ensure that multiple messages
// can be handled simultaneously.
func TestBus_Publish_Concurrent(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	received := make(chan *events.Content, 100)

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		received <- c
		return nil
	})

	for range 100 {
		content := &events.Content{
			URL:  "http://test.com",
			Type: events.TypeArticle,
		}
		err := bus.Publish(context.Background(), content)
		require.NoError(t, err)
	}

	for range 100 {
		select {
		case <-received:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for content")
		}
	}
}

// TestBus_Publish_ContextCancellation tests that context cancellation is respected
// during publishing.
func TestBus_Publish_ContextCancellation(t *testing.T) {
	t.Parallel()

	bus := events.NewBus()
	content := &events.Content{
		URL:  "http://test.com",
		Type: events.TypeArticle,
	}
	handlerDone := make(chan struct{})

	bus.Subscribe(func(ctx context.Context, c *events.Content) error {
		<-ctx.Done()
		close(handlerDone)
		return ctx.Err()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := bus.Publish(ctx, content)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)

	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for handler to finish")
	}
}
