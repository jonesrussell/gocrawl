package crawler_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	return &config.CrawlerConfig{
		MaxDepth:    3,
		Parallelism: 2,
	}
}

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	api.IndexManager
}

func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockIndex := &mockIndexManager{}

	// Verify mockLogger implements interfaces.Logger
	var _ interfaces.Logger = mockLogger

	app := fxtest.New(t,
		fx.Provide(
			func() interfaces.Logger { return mockLogger },
			func() config.Interface { return mockCfg },
			func() api.IndexManager { return mockIndex },
		),
		crawler.Module,
	)
	require.NoError(t, app.Err())
}

func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockIndex := &mockIndexManager{}

	// Verify mockLogger implements interfaces.Logger
	var _ interfaces.Logger = mockLogger

	var crawlerInstance crawler.Interface

	app := fxtest.New(t,
		crawler.Module,
		fx.Provide(
			func() interfaces.Logger { return mockLogger },
			func() config.Interface { return mockCfg },
			func() api.IndexManager { return mockIndex },
		),
		fx.Populate(&crawlerInstance),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, crawlerInstance)
}
