package crawler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
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

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	api.SearchManager
}

func (m *mockSearchManager) Search(_ context.Context, _ string, _ any) ([]any, error) {
	return []any{}, nil
}

func (m *mockSearchManager) Count(_ context.Context, _ string, _ any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Aggregate(_ context.Context, _ string, _ any) (any, error) {
	return nil, errors.New("aggregate not implemented in mock")
}

// mockIndexManager implements api.IndexManager for testing
type mockIndexManager struct {
	api.IndexManager
}

func (m *mockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockIndexManager) Close() error {
	return nil
}

// mockStorage implements types.Interface for testing
type mockStorage struct {
	types.Interface
}

func (m *mockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

// mockSources implements sources.Interface for testing
type mockSources struct {
	sources.Interface
}

func (m *mockSources) GetSource(_ string) (*sources.Config, error) {
	return &sources.Config{
		Name:      "test-source",
		URL:       "http://test.example.com",
		RateLimit: "1s",
		MaxDepth:  2,
		Selectors: sources.SelectorConfig{
			Title:       "h1",
			Description: "meta[name=description]",
			Content:     "article",
			Article: sources.ArticleSelectors{
				Container: "article",
				Title:     "h1",
				Body:      "article",
			},
		},
	}, nil
}

func (m *mockSources) ListSources() ([]*sources.Config, error) {
	return []*sources.Config{
		{
			Name:      "test-source",
			URL:       "http://test.example.com",
			RateLimit: "1s",
			MaxDepth:  2,
			Selectors: sources.SelectorConfig{
				Title:       "h1",
				Description: "meta[name=description]",
				Content:     "article",
				Article: sources.ArticleSelectors{
					Container: "article",
					Title:     "h1",
					Body:      "article",
				},
			},
		},
	}, nil
}

// mockSignalHandler implements signal.SignalHandler for testing
type mockSignalHandler struct {
	*signal.SignalHandler
}

func (m *mockSignalHandler) Setup(_ context.Context) func() {
	return func() {}
}

func (m *mockSignalHandler) SetLogger(_ logger.Interface) {}

func (m *mockSignalHandler) SetFXApp(_ *fx.App) {}

func (m *mockSignalHandler) Wait() bool {
	return true
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
				func() types.Interface {
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
	mockIndex := &mockIndexManager{}

	// Set up debug logging expectations
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	app := fxtest.New(t,
		fx.Supply(mockLogger, mockCfg),
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.ResultTags(`name:"testLogger"`),
			),
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.ResultTags(`name:"testConfig"`),
			),
			fx.Annotate(
				func() types.Interface { return mockStore },
				fx.As(new(types.Interface)),
			),
			fx.Annotate(
				func() api.IndexManager { return mockIndex },
				fx.As(new(api.IndexManager)),
			),
		),
		crawler.Module,
	)

	app.RequireStart()
	app.RequireStop()
}

// TestAppDependencies tests that all dependencies are properly wired up
func TestAppDependencies(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockCfg := &mockConfig{}
	mockStore := &mockStorage{}
	mockIndex := &mockIndexManager{}
	mockSearchManager := &mockSearchManager{}
	mockSources := &mockSources{}
	mockSignalHandler := signal.NewSignalHandler(mockLogger)

	// Set up debug logging expectations
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	app := fxtest.New(t,
		fx.NopLogger,
		crawler.Module,
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return mockLogger },
				fx.ResultTags(`name:"testLogger"`),
			),
			fx.Annotate(
				func() config.Interface { return mockCfg },
				fx.ResultTags(`name:"testConfig"`),
			),
			fx.Annotate(
				func() types.Interface { return mockStore },
				fx.ResultTags(`name:"testStorage"`),
			),
			fx.Annotate(
				func() api.IndexManager { return mockIndex },
				fx.ResultTags(`name:"testIndexManager"`),
			),
			fx.Annotate(
				func() api.SearchManager { return mockSearchManager },
				fx.ResultTags(`name:"testSearchManager"`),
			),
			fx.Annotate(
				func() sources.Interface { return mockSources },
				fx.ResultTags(`name:"sourceManager"`),
			),
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return mockSignalHandler },
				fx.ResultTags(`name:"signalHandler"`),
			),
			context.Background,
		),
		fx.Invoke(func(deps crawler.CrawlDeps) {
			assert.NotNil(t, deps.Logger)
			assert.NotNil(t, deps.Config)
			assert.NotNil(t, deps.Storage)
			assert.NotNil(t, deps.Crawler)
			assert.NotNil(t, deps.Processors)
			assert.NotNil(t, deps.Done)
			assert.NotNil(t, deps.Context)
			assert.NotNil(t, deps.Sources)
			assert.NotNil(t, deps.SourceName)
			assert.NotNil(t, deps.Handler)
		}),
	)

	app.RequireStart()
	app.RequireStop()
}
