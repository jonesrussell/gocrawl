package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
		fx.Provide(
			fx.Annotate(
				func() types.Logger { return mockLogger },
				fx.As(new(types.Logger)),
			),
		),
		fx.Supply(
			mockStore,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
		fx.Invoke(func(p indices.CreateParams) error {
			// Create the index
			if err := p.Storage.CreateIndex(p.Context, "test-index", nil); err != nil {
				return err
			}

			p.Logger.Info("Successfully created index", "name", "test-index")
			return nil
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestCreateCommandError tests error handling in the create index command
func TestCreateCommandError(t *testing.T) {
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
	mockStore.On("CreateIndex", mock.Anything, "test-index", mock.Anything).Return(storage.ErrIndexNotFound)

	// Create test app with mocked dependencies
	app := fxtest.New(t,
		fx.NopLogger,
		fx.Provide(
			fx.Annotate(
				func() types.Logger { return mockLogger },
				fx.As(new(types.Logger)),
			),
		),
		fx.Supply(
			mockStore,
			mockCfg,
		),
		fx.Provide(
			func() context.Context { return t.Context() },
			func() storagetypes.Interface { return mockStore },
			func() config.Interface { return mockCfg },
		),
	)

	// Start the app and verify it starts without errors
	err := app.Start(t.Context())
	require.NoError(t, err)
	defer app.RequireStop()

	// Get the command and run it
	cmd := indices.Command()
	cmd.SetArgs([]string{"create", "test-index"})
	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "error starting application")

	// Verify that the index creation was attempted
	mockStore.AssertCalled(t, "CreateIndex", mock.Anything, "test-index", mock.Anything)
}
