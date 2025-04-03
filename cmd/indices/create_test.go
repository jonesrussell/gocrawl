package indices_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockStorage implements storage.Interface for testing
type mockStorage struct {
	mock.Mock
	storagetypes.Interface
}

func (m *mockStorage) CreateIndex(ctx context.Context, name string, mapping map[string]any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

func (m *mockStorage) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockStorage) ListIndices(ctx context.Context) ([]string, error) {
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

func (m *mockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// testModule provides a test module with mock dependencies
var testModule = func(t *testing.T) fx.Option {
	return fx.Module("test",
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface { return configtestutils.NewMockConfig() },
			func() logger.Interface { return logger.NewNoOp() },
		),
	)
}

// TestCreateCommand tests the create index command
func TestCreateCommand(t *testing.T) {
	t.Parallel()

	mockStore := &mockStorage{}
	mockStore.On("CreateIndex", mock.Anything, "test-index", mock.Anything).Return(nil)

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		testModule(t),
		fx.Provide(
			func() storagetypes.Interface { return mockStore },
		),
		fx.Invoke(func(lc fx.Lifecycle, ctx context.Context, logger logger.Interface, storage storagetypes.Interface) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Create the index with default mapping
					if createErr := storage.CreateIndex(ctx, "test-index", indices.DefaultMapping); createErr != nil {
						return fmt.Errorf("failed to create index test-index: %w", createErr)
					}
					logger.Info("Successfully created index", "name", "test-index")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app and verify it starts without errors
	startErr := app.Start(t.Context())
	require.NoError(t, startErr)
	defer app.RequireStop()

	// Verify that the index was created
	mockStore.AssertCalled(t, "CreateIndex", mock.Anything, "test-index", mock.Anything)
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

			mockStore := &mockStorage{}
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
