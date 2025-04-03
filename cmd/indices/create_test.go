// Package indices_test provides tests for the indices command.
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
	"github.com/stretchr/testify/assert"
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

func (m *mockStorage) IndexExists(ctx context.Context, name string) (bool, error) {
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
			logger.NewNoOp,
		),
	)
}

func TestCreateCommand(t *testing.T) {
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
			mockStore := &mockStorage{}

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
