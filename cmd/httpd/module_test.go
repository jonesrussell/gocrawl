// Package httpd_test implements tests for the HTTP server command.
package httpd_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestModuleProvides tests that the HTTPD module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	t.Parallel()

	app := fxtest.New(t,
		fx.NopLogger,
		httpd.Module,
		fx.Provide(
			// Provide a mock storage interface
			func() types.Interface {
				return &httpdTestStorage{}
			},
		),
	)

	require.NoError(t, app.Err())
}

// httpdTestStorage implements types.Interface for testing
type httpdTestStorage struct {
	types.Interface
}
