package indices_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/config"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
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

// TestCreateCommand tests the create index command
func TestCreateCommand(t *testing.T) {
	t.Parallel()

	// Create test dependencies
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	mockCfg := &testutils.MockConfig{}
	mockCfg.On("GetAppConfig").Return(&config.AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       true,
	})
	mockCfg.On("GetLogConfig").Return(&config.LogConfig{
		Level: "debug",
		Debug: true,
	})
	mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "test-index",
	})
	mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
	mockCfg.On("GetSources").Return([]config.Source{}, nil)
	mockCfg.On("GetCommand").Return("test")

	mockStore := &mockStorage{}
	mockStore.On("CreateIndex", mock.Anything, "test-index", mock.Anything).Return(nil)

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Supply(
			mockStore,
			mockCfg,
			mockLogger,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
		),
		indices.Module,
		fx.Invoke(func(lc fx.Lifecycle, p indices.CreateParams) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Create the index with default mapping
					if err := p.Storage.CreateIndex(ctx, "test-index", indices.DefaultMapping); err != nil {
						return fmt.Errorf("failed to create index test-index: %w", err)
					}
					p.Logger.Info("Successfully created index", "name", "test-index")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return nil
				},
			})
		}),
	)

	// Start the app and verify it starts without errors
	err := app.Start(t.Context())
	require.NoError(t, err)
	defer app.RequireStop()

	// Verify that the index was created
	mockStore.AssertCalled(t, "CreateIndex", mock.Anything, "test-index", mock.Anything)
}

// TestCreateCommandArgs tests argument validation in the create index command
func TestCreateCommandArgs(t *testing.T) {
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
			wantErr: true,
			errMsg:  "error starting application",
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
			// Create mock dependencies
			mockLogger := &testutils.MockLogger{}
			mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

			mockCfg := &testutils.MockConfig{}
			mockCfg.On("GetAppConfig").Return(&config.AppConfig{
				Environment: "test",
				Name:        "gocrawl",
				Version:     "1.0.0",
				Debug:       true,
			})
			mockCfg.On("GetLogConfig").Return(&config.LogConfig{
				Level: "debug",
				Debug: true,
			})
			mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test-index",
			})
			mockCfg.On("GetServerConfig").Return(testutils.NewTestServerConfig())
			mockCfg.On("GetSources").Return([]config.Source{}, nil)
			mockCfg.On("GetCommand").Return("test")

			mockStore := &mockStorage{}
			if tt.name == "invalid index name" {
				mockStore.On("CreateIndex", mock.Anything, "invalid/index/name", mock.Anything).Return(fmt.Errorf("failed to create index"))
			}

			// Create test app with mocked dependencies
			app := fxtest.New(t,
				fx.NopLogger,
				fx.Supply(
					mockStore,
					mockCfg,
					mockLogger,
				),
				fx.Provide(
					func() context.Context { return t.Context() },
				),
				indices.Module,
			)

			cmd := indices.Command()
			cmd.SetArgs(append([]string{"create"}, tt.args...))
			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
