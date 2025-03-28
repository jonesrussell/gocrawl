// Package httpd_test implements tests for the HTTP server command.
package httpd_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtest "github.com/jonesrussell/gocrawl/internal/config/testutils"
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
			// Provide context for api.Module
			func() context.Context {
				return t.Context()
			},
			// Provide a mock config
			func() config.Interface {
				return configtest.NewMockConfig()
			},
		),
	)

	require.NoError(t, app.Err())
}

// httpdTestStorage implements types.Interface for testing
type httpdTestStorage struct {
	types.Interface
}
