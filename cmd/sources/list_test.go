// Package sources_test provides tests for the sources command package.
package sources_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	cmdsrcs "github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// ErrSourceNotFound is returned when a source is not found
var ErrSourceNotFound = errors.New("source not found")

// mockLogger implements logger.Interface for testing
type mockLogger struct {
	mock.Mock
	logger.Interface
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
	sources []sources.Config
	err     error
}

func (m *mockSourceManager) GetSources() ([]sources.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sources, nil
}

func (m *mockSourceManager) FindByName(name string) (*sources.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.sources {
		if m.sources[i].Name == name {
			return &m.sources[i], nil
		}
	}
	return nil, ErrSourceNotFound
}

func (m *mockSourceManager) Validate(_ *sources.Config) error { return nil }

func (m *mockSourceManager) ListSources(ctx context.Context) ([]*sources.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*sources.Config, len(m.sources))
	for i := range m.sources {
		result[i] = &m.sources[i]
	}
	return result, nil
}

func (m *mockSourceManager) AddSource(ctx context.Context, source *sources.Config) error {
	if m.err != nil {
		return m.err
	}
	m.sources = append(m.sources, *source)
	return nil
}

func (m *mockSourceManager) UpdateSource(ctx context.Context, source *sources.Config) error {
	if m.err != nil {
		return m.err
	}
	for i := range m.sources {
		if m.sources[i].Name == source.Name {
			m.sources[i] = *source
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

func (m *mockSourceManager) DeleteSource(ctx context.Context, name string) error {
	if m.err != nil {
		return m.err
	}
	for i := range m.sources {
		if m.sources[i].Name == name {
			m.sources = append(m.sources[:i], m.sources[i+1:]...)
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

func (m *mockSourceManager) ValidateSource(source *sources.Config) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockSourceManager) GetMetrics() sources.Metrics {
	return sources.Metrics{
		SourceCount: int64(len(m.sources)),
		LastUpdated: time.Now(),
	}
}

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
	Sources sources.Interface
	Logger  logger.Interface
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
	writeErr := os.WriteFile(filepath.Join(tmpDir, "sources.yml"), []byte(sourcesYml), 0644)
	require.NoError(t, writeErr)

	// Set environment variables for testing
	t.Setenv("SOURCES_FILE", filepath.Join(tmpDir, "sources.yml"))
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*cobra.Command, sources.Interface, logger.Interface)
		wantErr bool
	}{
		{
			name: "successful execution",
			setup: func(t *testing.T) (*cobra.Command, sources.Interface, logger.Interface) {
				cmd := &cobra.Command{}
				cmd.SetContext(t.Context())

				// Create mock logger
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()

				// Create mock source manager
				sm := &mockSourceManager{
					sources: []sources.Config{
						{
							Name:      "Test Source",
							URL:       "https://test.com",
							RateLimit: time.Second,
							MaxDepth:  2,
						},
					},
				}

				return cmd, sm, ml
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, sm, ml := tt.setup(t)

			// Create a test application with required modules
			app := fx.New(
				fx.NopLogger,
				fx.Supply(cmd),
				fx.Supply([]string{}),
				fx.Provide(
					fx.Annotate(
						func() sources.Interface { return sm },
						fx.As(new(sources.Interface)),
					),
					fx.Annotate(
						func() logger.Interface { return ml },
						fx.As(new(logger.Interface)),
					),
				),
				fx.Invoke(func(p struct {
					fx.In
					Sources sources.Interface
					Logger  logger.Interface
					LC      fx.Lifecycle
				}) {
					p.LC.Append(fx.Hook{
						OnStart: func(context.Context) error {
							params := &cmdsrcs.Params{
								SourceManager: p.Sources,
								Logger:        p.Logger,
							}
							if err := cmdsrcs.ExecuteList(*params, t.Context()); err != nil {
								p.Logger.Error("Error executing list", "error", err)
								return err
							}
							return nil
						},
						OnStop: func(context.Context) error {
							return nil
						},
					})
				}),
			)

			// Start the application
			startErr := app.Start(t.Context())
			if tt.wantErr {
				require.Error(t, startErr)
				return
			}
			require.NoError(t, startErr)

			// Stop the application
			stopErr := app.Stop(t.Context())
			require.NoError(t, stopErr)
		})
	}
}

func Test_executeList(t *testing.T) {
	t.Parallel()
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
					sources: []sources.Config{
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
					sources: []sources.Config{},
				}

				mc := newMockConfig()
				return ml, sm, mc
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ml, sm, mc := tt.setup(t)

			app := fxtest.New(t,
				fx.Supply(t.Context()),
				fx.Provide(
					func() sources.Interface { return sm },
					func() logger.Interface { return ml },
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
							if err := cmdsrcs.ExecuteList(*params, t.Context()); err != nil {
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

	tests := []struct {
		name    string
		setup   func(t *testing.T) cmdsrcs.Params
		wantErr bool
	}{
		{
			name: "error getting sources",
			setup: func(t *testing.T) cmdsrcs.Params {
				ml := &mockLogger{}
				ml.On("Info", mock.Anything, mock.Anything).Return()
				ml.On("Error", mock.Anything, mock.Anything).Return()
				ml.On("Debug", mock.Anything, mock.Anything).Return()
				ml.On("Warn", mock.Anything, mock.Anything).Return()
				ml.On("Fatal", mock.Anything, mock.Anything).Return()
				ml.On("Printf", mock.Anything, mock.Anything).Return()
				ml.On("Errorf", mock.Anything, mock.Anything).Return()
				ml.On("Sync").Return(nil)

				sm := &mockSourceManager{
					err: errors.New("failed to get sources"),
				}

				return cmdsrcs.Params{
					SourceManager: sm,
					Logger:        ml,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := tt.setup(t)
			err := cmdsrcs.ExecuteList(params, t.Context())
			if tt.wantErr {
				require.Error(t, err)
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
		sources []sources.Config
		logger  logger.Interface
		wantErr bool
	}{
		{
			name: "invalid source data",
			sources: []sources.Config{
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockLog := tt.logger.(*mockLogger)
			mockLog.On("Info", "Found sources", []any{"count", 1}).Return()
			mockLog.On("Info", "Source", []any{"name", "", "url", ""}).Return()

			err := cmdsrcs.PrintSources(tt.sources, tt.logger)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			mockLog.AssertExpectations(t)
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
	testConfigs := []sources.Config{
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

func TestModuleProvides(t *testing.T) {
	ml := &mockLogger{}

	app := fxtest.New(t,
		fx.Supply(ml),
		fx.Provide(
			fx.Annotate(func() logger.Interface { return ml }, fx.As(new(logger.Interface))),
		),
		cmdsrcs.Module,
	)

	app.RequireStart()
	app.RequireStop()
}
