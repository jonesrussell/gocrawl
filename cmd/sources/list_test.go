// Package sources_test provides tests for the sources command package.
package sources_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	cmdsrcs "github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
	internal "github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// ErrSourceNotFound is returned when a source is not found
var ErrSourceNotFound = errors.New("source not found")

// mockLogger implements common.Logger for testing
type mockLogger struct {
	mock.Mock
	common.Logger
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
	m.errorMessages = append(m.errorMessages, msg)
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

// mockSourceManager implements sources.Interface for testing
type mockSourceManager struct {
	sources []internal.Config
	err     error
}

func (m *mockSourceManager) GetSources() ([]internal.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func (m *mockSourceManager) FindByName(name string) (*internal.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.sources {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, ErrSourceNotFound
}

func (m *mockSourceManager) Validate(_ *internal.Config) error { return nil }

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
	sources []config.Source
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		sources: []config.Source{
			{
				Name:      "Test Source",
				URL:       "https://test.com",
				RateLimit: time.Second,
				MaxDepth:  2,
			},
		},
	}
}

func (m *mockConfig) GetString(_ string) string               { return "" }
func (m *mockConfig) GetStringSlice(_ string) []string        { return nil }
func (m *mockConfig) GetInt(_ string) int                     { return 0 }
func (m *mockConfig) GetBool(_ string) bool                   { return false }
func (m *mockConfig) GetDuration(_ string) time.Duration      { return 0 }
func (m *mockConfig) UnmarshalKey(_ string, _ any) error      { return nil }
func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig { return &config.CrawlerConfig{} }
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{}
}
func (m *mockConfig) GetLogConfig() *config.LogConfig {
	return &config.LogConfig{
		Level: "info",
	}
}
func (m *mockConfig) GetAppConfig() *config.AppConfig {
	return &config.AppConfig{
		Environment: "test",
		Debug:       false,
	}
}
func (m *mockConfig) GetServerConfig() *config.ServerConfig { return &config.ServerConfig{} }
func (m *mockConfig) GetSources() []config.Source           { return m.sources }
func (m *mockConfig) GetCommand() string                    { return "list" }

// TestParams holds the dependencies required for the list operation.
type TestParams struct {
	fx.In
	Sources internal.Interface
	Logger  types.Logger
}

// TestConfigModule provides test configuration
var TestConfigModule = fx.Module("test_config",
	fx.Provide(
		fx.Annotate(
			newMockConfig,
			fx.As(new(config.Interface)),
		),
	),
)

// TestCommonModule provides common test dependencies
var TestCommonModule = fx.Module("test_common",
	fx.Provide(
		context.Background,
	),
)

func Test_listCommand(t *testing.T) {
	t.Parallel()
	cmd := cmdsrcs.ListCommand()
	require.NotNil(t, cmd)
	require.Equal(t, "list", cmd.Use)
	require.Equal(t, "List all configured content sources", cmd.Short)
	require.Contains(t, cmd.Long, "Display a list of all content sources configured in sources.yml")
}

func Test_runList(t *testing.T) {
	// Create a temporary sources.yml file for testing
	tmpDir := t.TempDir()
	sourcesYml := `sources:
  - name: Test Source
    url: https://test.com
    rate_limit: 1s
    max_depth: 2
    index: test_content
    article_index: test_articles
`
	err := os.WriteFile(filepath.Join(tmpDir, "sources.yml"), []byte(sourcesYml), 0644)
	require.NoError(t, err)

	// Set environment variables for testing
	t.Setenv("SOURCES_FILE", filepath.Join(tmpDir, "sources.yml"))
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*cobra.Command, internal.Interface, types.Logger)
		wantErr bool
	}{
		{
			name: "successful execution",
			setup: func(t *testing.T) (*cobra.Command, internal.Interface, types.Logger) {
				cmd := &cobra.Command{}
				cmd.SetContext(context.Background())

				// Create mock logger
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()
				ml.On("Sync").Return(nil)

				// Create mock source manager
				sourceConfigs := []internal.Config{
					{
						Name:         "Test Source",
						URL:          "https://test.com",
						Index:        "test_content",
						ArticleIndex: "test_articles",
						RateLimit:    time.Second,
						MaxDepth:     2,
					},
				}
				sm := &mockSourceManager{sources: sourceConfigs}

				return cmd, sm, ml
			},
			wantErr: false,
		},
		{
			name: "context cancellation",
			setup: func(t *testing.T) (*cobra.Command, internal.Interface, types.Logger) {
				ctx, cancel := context.WithCancel(context.Background())
				cmd := &cobra.Command{}
				cmd.SetContext(ctx)

				// Create mock logger
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()
				ml.On("Sync").Return(nil)

				// Create mock source manager
				sourceConfigs := []internal.Config{
					{
						Name:         "Test Source",
						URL:          "https://test.com",
						Index:        "test_content",
						ArticleIndex: "test_articles",
						RateLimit:    time.Second,
						MaxDepth:     2,
					},
				}
				sm := &mockSourceManager{sources: sourceConfigs}

				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()

				return cmd, sm, ml
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd, sm, ml := tt.setup(t)

			// Create test app with all required dependencies
			app := fx.New(
				fx.NopLogger,
				fx.Supply(cmd),
				fx.Provide(
					fx.Annotate(
						func() types.Logger { return ml },
						fx.As(new(types.Logger)),
					),
					fx.Annotate(
						func() internal.Interface { return sm },
						fx.As(new(internal.Interface)),
					),
					fx.Annotate(
						newMockConfig,
						fx.As(new(config.Interface)),
					),
				),
				fx.Invoke(func(p struct {
					fx.In
					Sources internal.Interface
					Logger  types.Logger
					LC      fx.Lifecycle
				}) {
					// Create signal handler with the real logger
					handler := signal.NewSignalHandler(p.Logger)
					cleanup := handler.Setup(cmd.Context())
					p.LC.Append(fx.Hook{
						OnStart: func(context.Context) error {
							params := &cmdsrcs.Params{
								SourceManager: p.Sources,
								Logger:        p.Logger,
							}
							if err := cmdsrcs.ExecuteList(*params); err != nil {
								p.Logger.Error("Error executing list", "error", err)
								return err
							}
							return nil
						},
						OnStop: func(context.Context) error {
							cleanup()
							return nil
						},
					})
				}),
			)

			err := app.Start(cmd.Context())
			if err != nil {
				t.Fatalf("failed to start app: %v", err)
			}
			defer app.Stop(cmd.Context())

			// Create a new app for RunList with the same dependencies
			app = fx.New(
				fx.NopLogger,
				fx.Supply(cmd),
				fx.Provide(
					fx.Annotate(
						func() types.Logger { return ml },
						fx.As(new(types.Logger)),
					),
					fx.Annotate(
						func() sources.Interface { return sm },
						fx.As(new(sources.Interface)),
					),
					fx.Annotate(
						newMockConfig,
						fx.As(new(config.Interface)),
					),
				),
				fx.Invoke(func(p struct {
					fx.In
					Sources sources.Interface
					Logger  types.Logger
					LC      fx.Lifecycle
				}) {
					// Create signal handler with the real logger
					handler := signal.NewSignalHandler(p.Logger)
					cleanup := handler.Setup(cmd.Context())
					p.LC.Append(fx.Hook{
						OnStart: func(context.Context) error {
							params := &cmdsrcs.Params{
								SourceManager: p.Sources,
								Logger:        p.Logger,
							}
							if err := cmdsrcs.ExecuteList(*params); err != nil {
								p.Logger.Error("Error executing list", "error", err)
								return err
							}
							return nil
						},
						OnStop: func(context.Context) error {
							cleanup()
							return nil
						},
					})
				}),
			)

			err = app.Start(cmd.Context())
			if err != nil {
				t.Fatalf("failed to start app: %v", err)
			}
			defer app.Stop(cmd.Context())

			// Wait for the app to start
			time.Sleep(100 * time.Millisecond)

			err = cmdsrcs.RunList(cmd, nil)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_executeList(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*mockLogger, *mockSourceManager, *mockConfig)
		wantErr bool
	}{
		{
			name: "with sources",
			setup: func(t *testing.T) (*mockLogger, *mockSourceManager, *mockConfig) {
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()
				ml.On("Sync").Return(nil)
				ml.On("Info", "No sources found", mock.Anything).Return()

				sm := &mockSourceManager{
					sources: []internal.Config{
						{
							Name:         "Test Source",
							URL:          "https://test.com",
							Index:        "test_content",
							ArticleIndex: "test_articles",
							RateLimit:    time.Second,
							MaxDepth:     2,
						},
					},
				}

				mc := newMockConfig()
				return ml, sm, mc
			},
			wantErr: false,
		},
		{
			name: "no sources",
			setup: func(t *testing.T) (*mockLogger, *mockSourceManager, *mockConfig) {
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()
				ml.On("Sync").Return(nil)
				ml.On("Info", "No sources found", mock.Anything).Return()

				sm := &mockSourceManager{
					sources: []internal.Config{},
				}

				mc := newMockConfig()
				return ml, sm, mc
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml, sm, mc := tt.setup(t)

			app := fxtest.New(t,
				fx.Supply(t.Context()),
				fx.Provide(
					func() internal.Interface { return sm },
					func() common.Logger { return ml },
					fx.Annotate(
						func() config.Interface { return mc },
						fx.As(new(config.Interface)),
					),
				),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							params := &cmdsrcs.Params{
								SourceManager: sm,
								Logger:        ml,
							}
							if err := cmdsrcs.ExecuteList(*params); err != nil {
								ml.Error("Error executing list", "error", err)
								return err
							}
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return nil
						},
					})
				}),
			)

			if tt.wantErr {
				require.Error(t, app.Err())
			} else {
				app.RequireStart()
				app.RequireStop()
			}
		})
	}
}

func Test_executeList_error(t *testing.T) {
	t.Parallel()

	errSourceManager := &errorSourceManager{
		err: errors.New("failed to get sources"),
	}

	ml := &mockLogger{}
	ml.On("Info", mock.Anything, mock.Anything).Return()
	ml.On("Error", mock.Anything, mock.Anything).Return()
	ml.On("Debug", mock.Anything, mock.Anything).Return()
	ml.On("Warn", mock.Anything, mock.Anything).Return()
	ml.On("Fatal", mock.Anything, mock.Anything).Return()
	ml.On("Printf", mock.Anything, mock.Anything).Return()
	ml.On("Errorf", mock.Anything, mock.Anything).Return()
	ml.On("Sync").Return(nil)

	tests := []struct {
		name    string
		params  cmdsrcs.Params
		wantErr bool
	}{
		{
			name: "error getting sources",
			params: cmdsrcs.Params{
				SourceManager: errSourceManager,
				Logger:        ml,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmdsrcs.ExecuteList(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get sources")
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_printSources_error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sources []internal.Config
		logger  types.Logger
		wantErr bool
	}{
		{
			name: "invalid source data",
			sources: []internal.Config{
				{
					Name:      "", // Empty name should cause formatting issues
					URL:       "",
					RateLimit: 0,
					MaxDepth:  -1,
				},
			},
			logger:  &mockLogger{},
			wantErr: false, // PrintSources should handle invalid data gracefully
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmdsrcs.PrintSources(tt.sources, tt.logger)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCommand(t *testing.T) {
	t.Parallel()
	cmd := cmdsrcs.Command()
	require.NotNil(t, cmd)
	require.Equal(t, "sources", cmd.Use)
	require.Equal(t, "Manage sources defined in sources.yml", cmd.Short)
	require.Contains(t, cmd.Long, "Manage sources defined in sources.yml")

	// Test subcommands
	subCmds := cmd.Commands()
	require.NotEmpty(t, subCmds)

	// Verify list subcommand is present
	var hasListCmd bool
	for _, subCmd := range subCmds {
		if subCmd.Name() == "list" {
			hasListCmd = true
			break
		}
	}
	require.True(t, hasListCmd, "list subcommand should be present")
}

func TestFindByName(t *testing.T) {
	t.Parallel()
	testConfigs := []internal.Config{
		{
			Name:      "test1",
			URL:       "https://example1.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
		{
			Name:      "test2",
			URL:       "https://example2.com",
			RateLimit: 2 * time.Second,
			MaxDepth:  2,
		},
	}
	s := &mockSourceManager{sources: testConfigs}
	require.NotNil(t, s)

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "existing source",
			source:  "test1",
			wantErr: false,
		},
		{
			name:    "non-existing source",
			source:  "test3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source, err := s.FindByName(tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.source, source.Name)
		})
	}
}

type errorSourceManager struct {
	mockSourceManager
	err error
}

func (m *errorSourceManager) GetSources() ([]internal.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}
