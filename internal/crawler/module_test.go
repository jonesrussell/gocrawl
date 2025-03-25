package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
		RateLimit:   time.Second,
	}
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		APIKey:    "test-api-key", // Required API key
		IndexName: "test-index",
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{
			ID:     "test-deployment",
			APIKey: "test-cloud-key",
		},
		TLS: struct {
			Enabled     bool   `yaml:"enabled"`
			SkipVerify  bool   `yaml:"skip_verify"`
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
			CA          string `yaml:"ca"`
		}{
			Enabled:    true,
			SkipVerify: true,
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     true,
			InitialWait: 1 * time.Second,
			MaxWait:     30 * time.Second,
			MaxRetries:  3,
		},
	}
}

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	api.IndexManager
}

// mockStorage implements storage.Interface for testing
type mockStorage struct {
	storage.Interface
	testConnectionFn func(ctx context.Context) error
	closeFn          func() error
}

func (m *mockStorage) TestConnection(ctx context.Context) error {
	if m.testConnectionFn != nil {
		return m.testConnectionFn(ctx)
	}
	return nil
}

func (m *mockStorage) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

// TestModule tests that the crawler module provides all necessary dependencies
func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockIndex := &mockIndexManager{}
	mockSearchManager := api.NewMockSearchManager(ctrl)
	mockStorage := &mockStorage{}

	// Set up debug logging expectations
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Verify mockLogger implements logger.Interface
	var _ logger.Interface = mockLogger

	app := fxtest.New(t,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface {
					mockLogger.Debug("Providing test logger")
					return mockLogger
				},
				fx.ResultTags(`name:"testLogger"`),
			),
			fx.Annotate(
				func() config.Interface {
					mockLogger.Debug("Providing test config")
					return mockCfg
				},
				fx.ResultTags(`name:"testConfig"`),
			),
			fx.Annotate(
				func() api.IndexManager {
					mockLogger.Debug("Providing test index manager")
					return mockIndex
				},
				fx.ResultTags(`name:"testIndexManager"`),
			),
			fx.Annotate(
				func() api.SearchManager {
					mockLogger.Debug("Providing test search manager")
					return mockSearchManager
				},
				fx.ResultTags(`name:"testSearchManager"`),
			),
			fx.Annotate(
				func() storage.Interface {
					mockLogger.Debug("Providing test storage")
					return mockStorage
				},
				fx.ResultTags(`name:"testStorage"`),
			),
		),
		crawler.Module,
	)
	require.NoError(t, app.Err())
}

// TestModuleProvides tests that the crawler module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockSearchManager := api.NewMockSearchManager(ctrl)

	// Set up debug logging expectations
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock storage methods
	mockStore.testConnectionFn = func(ctx context.Context) error {
		return nil
	}
	mockStore.closeFn = func() error {
		return nil
	}

	// Verify mockLogger implements logger.Interface
	var _ logger.Interface = mockLogger

	var crawlerInstance crawler.Interface

	app := fxtest.New(t,
		fx.Supply(mockLogger),
		fx.Supply(mockCfg),
		fx.Supply(mockStore),
		fx.Supply(mockSearchManager),
		fx.Provide(
			fx.Annotate(
				func() logger.Interface {
					return mockLogger
				},
				fx.ResultTags(`name:"testLogger"`),
			),
			fx.Annotate(
				func() config.Interface {
					return mockCfg
				},
				fx.ResultTags(`name:"testConfig"`),
			),
			fx.Annotate(
				func() storage.Interface {
					return mockStore
				},
				fx.ResultTags(`name:"testStorage"`),
			),
			fx.Annotate(
				func() api.SearchManager {
					return mockSearchManager
				},
				fx.ResultTags(`name:"testSearchManager"`),
			),
			fx.Annotate(
				func() api.IndexManager {
					return &mockIndexManager{}
				},
				fx.ResultTags(`name:"testIndexManager"`),
			),
		),
		crawler.Module,
		fx.Populate(&crawlerInstance),
	)

	require.NoError(t, app.Err())
	app.RequireStart()
	defer app.RequireStop()

	require.NotNil(t, crawlerInstance)
}
