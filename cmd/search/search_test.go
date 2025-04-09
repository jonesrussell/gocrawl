// Package search_test implements tests for the search command.
package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	mock.Mock
}

func (m *mockSearchManager) Search(ctx context.Context, index string, query map[string]any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, errors.New("invalid search result type")
	}
	return val, nil
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, errors.New("invalid count result type")
	}
	return val, nil
}

func (m *mockSearchManager) Aggregate(ctx context.Context, index string, aggs map[string]any) (map[string]any, error) {
	args := m.Called(ctx, index, aggs)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, errors.New("invalid aggregation result type")
	}
	return val, nil
}

func (m *mockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// testDeps holds all the dependencies needed for testing
type testDeps struct {
	Storage       types.Interface
	Logger        logger.Interface
	Config        config.Interface
	Handler       *signal.SignalHandler
	Context       context.Context
	SearchManager api.SearchManager
}

// setupTestDeps creates and configures all test dependencies
func setupTestDeps(t *testing.T) *testDeps {
	t.Helper()

	// Create mock dependencies
	mockLogger := testutils.NewMockLogger()
	mockStorage := testutils.NewMockStorage(mockLogger)
	mockHandler := signal.NewSignalHandler(mockLogger)
	mockConfig := configtestutils.MockConfig{}
	mockConfig.On("GetAppConfig").Return(&app.Config{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockSearchManager := &mockSearchManager{}

	return &testDeps{
		Storage:       mockStorage,
		Logger:        mockLogger,
		Config:        &mockConfig,
		Handler:       mockHandler,
		Context:       t.Context(),
		SearchManager: mockSearchManager,
	}
}

// createTestApp creates a test application with the given dependencies
func createTestApp(t *testing.T, deps *testDeps) *fx.App {
	t.Helper()

	providers := []any{
		// Core dependencies
		func() types.Interface { return deps.Storage },
		func() logger.Interface { return deps.Logger },
		func() config.Interface { return deps.Config },
		func() *signal.SignalHandler { return deps.Handler },
		func() context.Context { return deps.Context },
		func() api.SearchManager { return deps.SearchManager },
		// Named parameters
		fx.Annotate(
			func() string { return "test-index" },
			fx.ResultTags(`name:"indexName"`),
		),
		fx.Annotate(
			func() string { return "test query" },
			fx.ResultTags(`name:"query"`),
		),
		fx.Annotate(
			func() int { return 10 },
			fx.ResultTags(`name:"resultSize"`),
		),
	}

	// Add command channels
	commandDone := make(chan struct{})
	providers = append(providers,
		func() chan struct{} { return commandDone },
	)

	// Create the app
	app := fx.New(
		fx.Provide(providers...),
		fx.Invoke(func(lc fx.Lifecycle, p search.Params) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Execute the search command
					if err := search.ExecuteSearch(ctx, p); err != nil {
						p.Logger.Error("Error executing search", "error", err)
						return err
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// Close channels
					close(commandDone)
					return nil
				},
			})
		}),
	)

	return app
}

// runTestApp runs the test app and handles cleanup
func runTestApp(t *testing.T, app *fx.App) {
	t.Helper()

	// Start the app
	err := app.Start(t.Context())
	require.NoError(t, err)

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	err = app.Stop(t.Context())
	require.NoError(t, err)
}

// TestCommandExecution tests the search command execution.
func TestCommandExecution(t *testing.T) {
	// Set up test dependencies
	deps := setupTestDeps(t)

	// Set up logger expectations
	deps.Logger.(*testutils.MockLogger).On("Info", "Starting search...", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	deps.Logger.(*testutils.MockLogger).On("Info", "Search completed", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Set up search manager expectations
	expectedQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": "test query",
			},
		},
		"size": 10,
	}
	deps.SearchManager.(*mockSearchManager).On("Search", mock.Anything, "test-index", expectedQuery).Return([]any{}, nil)
	deps.SearchManager.(*mockSearchManager).On("Close").Return(nil)

	// Create and run test app
	app := createTestApp(t, deps)
	runTestApp(t, app)

	// Verify mock expectations
	deps.SearchManager.(*mockSearchManager).AssertExpectations(t)
	deps.Logger.(*testutils.MockLogger).AssertExpectations(t)
}

func TestSearchCommand(t *testing.T) {
	// Create mock dependencies
	mockLogger := testutils.NewMockLogger()
	mockStorage := testutils.NewMockStorage(mockLogger)
	mockStorageMock := mockStorage.(*testutils.MockStorage)
	mockStorageMock.On("TestConnection", mock.Anything).Return(nil)

	mockConfig := &configtestutils.MockConfig{}
	mockConfig.On("GetAppConfig").Return(&config.AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockConfig.On("GetLogConfig").Return(&config.LogConfig{
		Level: "debug",
		Debug: true,
	})
	mockConfig.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	})
	mockConfig.On("GetServerConfig").Return(&config.ServerConfig{
		Address: ":8080",
	})
	mockConfig.On("GetSources").Return([]config.Source{}, nil)
	mockConfig.On("GetCommand").Return("test")
	mockConfig.On("GetPriorityConfig").Return(&config.PriorityConfig{
		Default: 1,
		Rules:   []config.PriorityRule{},
	})

	cmd := search.Command()
	require.NotNil(t, cmd)
	require.Equal(t, "search", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.NotEmpty(t, cmd.Long)
}
