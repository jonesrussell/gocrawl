// Package httpd_test implements tests for the HTTP server command.
package httpd_test

import (
	"context"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestHTTPServerCommand tests the HTTP server command initialization and shutdown
func TestHTTPServerCommand(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "HTTP server started", "address", ":8080").Return()
	mockLogger.On("Info", "Shutting down HTTP server...").Return()

	// Create a test server
	testServer := &http.Server{
		Addr: ":8080",
	}

	var params httpd.Params

	// Create test app
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() *http.Server { return testServer },
		),
		fx.Populate(&params),
	)

	// Start the app
	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Verify dependencies were provided
	assert.NotNil(t, params.Logger, "Logger should be injected into params")
	assert.NotNil(t, params.Server, "Server should be injected into params")
	assert.Equal(t, ":8080", params.Server.Addr, "Server address should match configuration")
}

// TestHTTPServerGracefulShutdown tests the graceful shutdown behavior
func TestHTTPServerGracefulShutdown(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "HTTP server started", "address", ":8080").Return()
	mockLogger.On("Info", "Shutting down HTTP server...").Return()

	// Create a test server
	testServer := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	// Create test app with server
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() *http.Server { return testServer },
		),
		httpd.Module,
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	// Start the app
	require.NoError(t, app.Start(ctx))

	// Wait a bit for the server to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app gracefully
	require.NoError(t, app.Stop(ctx))

	// Verify logger expectations
	mockLogger.AssertExpectations(t)
}

// TestHTTPServerSignalHandling tests the server's response to system signals
func TestHTTPServerSignalHandling(t *testing.T) {
	// Create test dependencies
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Info", "HTTP server started", "address", ":8080").Return()
	mockLogger.On("Info", "Shutting down HTTP server...").Return()

	// Create a test server
	testServer := &http.Server{
		Addr: ":8080",
	}

	// Get the command
	cmd := httpd.Command()
	assert.NotNil(t, cmd)

	// Create a context that we can cancel
	ctx, cancel := context.WithTimeout(t.Context(), common.DefaultShutdownTimeout)
	defer cancel()

	// Set up command context
	cmd.SetContext(ctx)

	// Create test app with server
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() *http.Server { return testServer },
		),
		httpd.Module,
	)

	// Start the app
	require.NoError(t, app.Start(ctx))
	defer app.Stop(ctx)

	// Start the command in a goroutine
	errChan := make(chan error)
	go func() {
		errChan <- cmd.Execute()
	}()

	// Wait a bit for the server to start
	time.Sleep(100 * time.Millisecond)

	// Simulate SIGTERM
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	// Wait for command to finish or timeout
	select {
	case err := <-errChan:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Command did not shut down within timeout")
	}

	// Verify logger expectations
	mockLogger.AssertExpectations(t)
}
