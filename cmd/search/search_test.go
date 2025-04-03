// Package search_test implements tests for the search command.
package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	mock.Mock
}

func (m *mockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]any), args.Error(1)
}

func (m *mockSearchManager) Aggregate(ctx context.Context, index string, query any) (any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0), args.Error(1)
}

func (m *mockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
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
	mockConfig := testutils.NewMockConfig()
	mockSearchManager := &mockSearchManager{}

	return &testDeps{
		Storage:       mockStorage,
		Logger:        mockLogger,
		Config:        mockConfig,
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
}

func TestSearchCommand(t *testing.T) {
	deps := setupTestDeps(t)
	app := createTestApp(t, deps)
	runTestApp(t, app)

	cmd := search.Command()
	assert.Equal(t, "search", cmd.Use)
	assert.Equal(t, "Search content in Elasticsearch", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Test flags
	err := cmd.Execute()
	require.Error(t, err, "Command should fail without required query flag")

	// Test flag defaults
	assert.Equal(t, "articles", cmd.Flag("index").DefValue)
	assert.Equal(t, "10", cmd.Flag("size").DefValue)
}
