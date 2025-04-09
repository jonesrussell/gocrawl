// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// MockStorage implements storage.Interface for testing
type MockStorage struct {
	mock.Mock
	storagetypes.Interface
}

func (m *MockStorage) CreateIndex(ctx context.Context, name string, mapping map[string]any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

func (m *MockStorage) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]string)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockStorage) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	if err := args.Error(1); err != nil {
		return false, err
	}
	val, ok := args.Get(0).(bool)
	if !ok {
		return false, nil
	}
	return val, nil
}

func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Add missing interface methods
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// testModule provides a test module with mock dependencies
var testModule = func(t *testing.T) fx.Option {
	return fx.Module("test",
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface {
				mockCfg := &configtestutils.MockConfig{}
				mockCfg.On("GetAppConfig").Return(&app.Config{
					Environment: "test",
					Name:        "gocrawl",
					Version:     "1.0.0",
					Debug:       true,
				})
				mockCfg.On("GetLogConfig").Return(&log.Config{
					Level: "debug",
				})
				mockCfg.On("GetElasticsearchConfig").Return(&elasticsearch.Config{
					Addresses: []string{"http://localhost:9200"},
					IndexName: "test-index",
				})
				mockCfg.On("GetServerConfig").Return(&server.Config{
					Address: ":8080",
				})
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				mockCfg.On("GetPriorityConfig").Return(&priority.Config{
					DefaultPriority: 1,
					Rules:           []priority.Rule{},
				})
				return mockCfg
			},
			logger.NewNoOp,
		),
	)
}

func TestCreateCommand(t *testing.T) {
	// Create mock dependencies
	mockLogger := testutils.NewMockLogger()
	mockStorage := testutils.NewMockStorage(mockLogger)
	mockStorageMock := mockStorage.(*testutils.MockStorage)
	mockStorageMock.On("TestConnection", mock.Anything).Return(nil)

	mockConfig := &configtestutils.MockConfig{}
	mockConfig.On("GetAppConfig").Return(&app.Config{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockConfig.On("GetLogConfig").Return(&log.Config{
		Level: "debug",
	})
	mockConfig.On("GetElasticsearchConfig").Return(&elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	})
	mockConfig.On("GetServerConfig").Return(&server.Config{
		Address: ":8080",
	})
	mockConfig.On("GetSources").Return([]config.Source{}, nil)
	mockConfig.On("GetCommand").Return("test")
	mockConfig.On("GetPriorityConfig").Return(&priority.Config{
		DefaultPriority: 1,
		Rules:           []priority.Rule{},
	})

	cmd := indices.CreateCommand()
	require.NotNil(t, cmd)
	require.Equal(t, "create [index-name]", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.NotEmpty(t, cmd.Long)

	t.Parallel()

	tests := []struct {
		name      string
		index     string
		exists    bool
		existsErr error
		createErr error
	}{
		{
			name:      "create new index",
			index:     "test-index",
			exists:    false,
			existsErr: nil,
			createErr: nil,
		},
		{
			name:      "index already exists",
			index:     "existing-index",
			exists:    true,
			existsErr: nil,
			createErr: nil,
		},
		{
			name:      "check existence error",
			index:     "error-index",
			exists:    false,
			existsErr: assert.AnError,
			createErr: nil,
		},
		{
			name:      "create error",
			index:     "create-error-index",
			exists:    false,
			existsErr: nil,
			createErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock storage
			mockStore := &MockStorage{}

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("IndexExists", mock.Anything, tt.index).Return(tt.exists, tt.existsErr)
			if !tt.exists && tt.existsErr == nil {
				mockStore.On("CreateIndex", mock.Anything, tt.index, indices.DefaultMapping).Return(tt.createErr)
			}

			// Create test app
			app := fxtest.New(t,
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() string { return tt.index },
					indices.NewCreator,
				),
				fx.Invoke(func(lc fx.Lifecycle, creator *indices.Creator, ctx context.Context) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := creator.Start(ctx)
							if tt.existsErr != nil || tt.createErr != nil {
								require.Error(t, err)
								if tt.existsErr != nil {
									require.ErrorIs(t, err, tt.existsErr)
								} else {
									require.ErrorIs(t, err, tt.createErr)
								}
							} else if tt.exists {
								require.Error(t, err)
								require.Contains(t, err.Error(), "already exists")
							} else {
								require.NoError(t, err)
							}
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return nil
						},
					})
				}),
			)

			// Start the app
			err := app.Start(t.Context())
			require.NoError(t, err)
			defer app.RequireStop()

			// Verify all expected calls were made
			mockStore.AssertExpectations(t)
		})
	}
}

// TestCreateCommandArgs tests argument validation in the create index command
func TestCreateCommandArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
			errMsg:  "accepts 1 arg(s)",
		},
		{
			name:    "too many args",
			args:    []string{"index1", "index2"},
			wantErr: true,
			errMsg:  "accepts 1 arg(s)",
		},
		{
			name:    "valid args",
			args:    []string{"index1"},
			wantErr: false,
		},
		{
			name:    "invalid index name",
			args:    []string{"invalid/index/name"},
			wantErr: true,
			errMsg:  "failed to create index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := &MockStorage{}
			switch tt.name {
			case "invalid index name":
				mockStore.On("CreateIndex", mock.Anything, "invalid/index/name", mock.Anything).
					Return(errors.New("failed to create index"))
			case "valid args":
				mockStore.On("CreateIndex", mock.Anything, "index1", mock.Anything).Return(nil)
			}

			cmd := indices.Command()
			cmd.SetArgs(append([]string{"create"}, tt.args...))

			// Execute the command to check argument validation
			execErr := cmd.Execute()
			if tt.name == "no args" || tt.name == "too many args" {
				require.Error(t, execErr)
				require.Contains(t, execErr.Error(), tt.errMsg)
				return
			}

			// For valid args and invalid index name, create test app with mocked dependencies
			app := fxtest.New(t,
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() *cobra.Command { return cmd },
				),
				fx.Invoke(func(lc fx.Lifecycle, ctx context.Context, logger logger.Interface, storage storagetypes.Interface) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							indexName := tt.args[0]
							if createErr := storage.CreateIndex(ctx, indexName, indices.DefaultMapping); createErr != nil {
								return fmt.Errorf("failed to create index %s: %w", indexName, createErr)
							}
							logger.Info("Successfully created index", "name", indexName)
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return nil
						},
					})
				}),
			)

			// Start the app
			startErr := app.Start(t.Context())
			if tt.wantErr {
				require.Error(t, startErr)
				require.Contains(t, startErr.Error(), tt.errMsg)
			} else {
				require.NoError(t, startErr)
			}
			app.RequireStop()
		})
	}
}
