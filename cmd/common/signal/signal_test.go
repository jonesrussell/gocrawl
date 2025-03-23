package signal_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalHandler(t *testing.T) {
	t.Parallel()

	t.Run("handles SIGINT", func(t *testing.T) {
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT in a goroutine
		go func() {
			time.Sleep(100 * time.Millisecond)
			p, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			require.NoError(t, p.Signal(syscall.SIGINT))
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Cancel context in a goroutine
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		// Wait for cancellation
		assert.True(t, handler.Wait())
	})

	t.Run("handles timeout", func(t *testing.T) {
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Create timeout context
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer timeoutCancel()

		// Wait with timeout
		assert.False(t, handler.WaitWithTimeout(timeoutCtx))
	})

	t.Run("calls cleanup function", func(t *testing.T) {
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cleanupCalled := false
		handler.SetCleanup(func() {
			cleanupCalled = true
		})

		cleanup := handler.Setup(ctx)
		cleanup()

		assert.True(t, cleanupCalled)
	})
}
