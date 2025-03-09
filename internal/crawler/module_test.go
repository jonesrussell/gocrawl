package crawler_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestModule(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockIndexManager := api.NewMockIndexManager()

	var c crawler.Interface

	app := fxtest.New(t,
		crawler.Module,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.IndexManager { return mockIndexManager },
		),
		fx.Populate(&c),
	)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))
	defer app.Stop(ctx)

	// Test that the crawler was provided and properly configured
	require.NotNil(t, c, "crawler should be provided")
	require.NotNil(t, c.GetIndexManager(), "index manager should be injected")
}
