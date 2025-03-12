package httpd_test

import (
	"net/http"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockLogger implements logger.Interface for testing
type mockLogger struct {
	logger.Interface
}

func TestModule(t *testing.T) {
	// Create mock dependencies
	mockLogger := &mockLogger{}

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
		),
		httpd.Module,
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())
	require.NoError(t, app.Err())
}

// TestModuleProvides tests that the httpd module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	var server *http.Server

	app := fxtest.New(t,
		httpd.Module,
		fx.Provide(
			func() logger.Interface { return &mockLogger{} },
			func() api.SearchManager { return &mockSearchManager{} },
		),
		fx.Populate(&server),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, server)
}

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}
