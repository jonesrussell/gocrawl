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
		t.Parallel()
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Send SIGINT in a goroutine
		var err error
		go func() {
			time.Sleep(100 * time.Millisecond)
			p, findErr := os.FindProcess(os.Getpid())
			if findErr != nil {
				err = findErr
				return
			}
			if sigErr := p.Signal(syscall.SIGINT); sigErr != nil {
				err = sigErr
				return
			}
		}()

		// Wait for signal
		assert.True(t, handler.Wait())
		require.NoError(t, err)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Parallel()
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(t.Context())
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
		t.Parallel()
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Create timeout context
		timeoutCtx, timeoutCancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer timeoutCancel()

		// Wait with timeout
		assert.False(t, handler.WaitWithTimeout(timeoutCtx))
	})

	t.Run("calls cleanup function", func(t *testing.T) {
		t.Parallel()
		handler := signal.NewSignalHandler()
		ctx, cancel := context.WithCancel(t.Context())
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
