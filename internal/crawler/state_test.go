package crawler_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	t.Parallel()

	t.Run("NewState", func(t *testing.T) {
		t.Parallel()
		s := crawler.NewState()
		assert.NotNil(t, s)
	})

	t.Run("StartStop", func(t *testing.T) {
		t.Parallel()
		s := crawler.NewState()
		ctx := t.Context()

		// Test Start
		s.Start(ctx, "test-source")
		assert.True(t, s.IsRunning())
		assert.Equal(t, "test-source", s.CurrentSource())
		assert.NotZero(t, s.StartTime())
		assert.NotNil(t, s.Context())

		// Test Stop
		s.Stop()
		assert.False(t, s.IsRunning())
		assert.Empty(t, s.CurrentSource())
		assert.Nil(t, s.Context())
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()
		s := crawler.NewState()
		ctx := t.Context()

		s.Start(ctx, "test-source")
		require.NotNil(t, s.Context())

		// Test Cancel
		s.Cancel()
		select {
		case <-s.Context().Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Fatal("context not cancelled")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		t.Parallel()
		s := crawler.NewState()

		// Start multiple goroutines to access state
		done := make(chan struct{})
		for range 10 {
			go func() {
				for range 100 {
					s.IsRunning()
					s.CurrentSource()
					s.StartTime()
					s.Context()
				}
				done <- struct{}{}
			}()
		}

		// Wait for all goroutines to complete
		for range 10 {
			<-done
		}
	})

	t.Run("StateTransitions", func(t *testing.T) {
		t.Parallel()
		s := crawler.NewState()
		ctx := t.Context()

		// Test multiple start/stop cycles
		for range 5 {
			source := "test-source"
			s.Start(ctx, source)
			assert.True(t, s.IsRunning())
			assert.Equal(t, source, s.CurrentSource())

			s.Stop()
			assert.False(t, s.IsRunning())
			assert.Empty(t, s.CurrentSource())
		}
	})
}
