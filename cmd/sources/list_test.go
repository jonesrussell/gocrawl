// Package sources_test provides tests for the sources command package.
package sources_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	srcs "github.com/jonesrussell/gocrawl/internal/sources"
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
	sources []srcs.Config
}

func (m *mockSourceManager) GetSources() ([]srcs.Config, error) { return m.sources, nil }

func (m *mockSourceManager) FindByName(name string) (*srcs.Config, error) {
	for _, s := range m.sources {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, ErrSourceNotFound
}

func (m *mockSourceManager) Validate(_ *srcs.Config) error { return nil }

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
func (m *mockConfig) GetLogConfig() *config.LogConfig       { return &config.LogConfig{} }
func (m *mockConfig) GetAppConfig() *config.AppConfig       { return &config.AppConfig{} }
func (m *mockConfig) GetServerConfig() *config.ServerConfig { return &config.ServerConfig{} }
func (m *mockConfig) GetSources() []config.Source           { return m.sources }

// TestParams holds the dependencies required for the list operation.
type TestParams struct {
	fx.In
	Sources srcs.Interface
	Logger  common.Logger
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
	cmd := sources.ListCommand()
	require.NotNil(t, cmd)
	require.Equal(t, "list", cmd.Use)
	require.Equal(t, "List all configured content sources", cmd.Short)
	require.Contains(t, cmd.Long, "Display a list of all content sources configured in sources.yml")
}

func Test_runList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*cobra.Command, *mockSourceManager, *mockLogger, *mockConfig)
		wantErr bool
	}{
		{
			name: "successful execution",
			setup: func(t *testing.T) (*cobra.Command, *mockSourceManager, *mockLogger, *mockConfig) {
				cmd := &cobra.Command{}
				cmd.SetContext(t.Context())
				sourceConfigs := []srcs.Config{
					{
						Name:         "Test Source",
						URL:          "https://test.com",
						Index:        "test_content",
						ArticleIndex: "test_articles",
						RateLimit:    time.Second,
						MaxDepth:     2,
					},
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
				return cmd, &mockSourceManager{sources: sourceConfigs}, ml, newMockConfig()
			},
			wantErr: false,
		},
		{
			name: "no sources found",
			setup: func(t *testing.T) (*cobra.Command, *mockSourceManager, *mockLogger, *mockConfig) {
				cmd := &cobra.Command{}
				cmd.SetContext(t.Context())
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
				return cmd, &mockSourceManager{sources: nil}, ml, newMockConfig()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, sm, ml, mc := tt.setup(t)

			// Create test app with mock dependencies
			app := fxtest.New(t,
				fx.Supply(cmd.Context()),
				fx.Provide(
					// Provide source manager without name tag
					func() srcs.Interface { return sm },
					func() common.Logger { return ml },
					// Provide config with the correct interface
					fx.Annotate(
						func() config.Interface { return mc },
						fx.As(new(config.Interface)),
					),
					// Provide cobra command and args
					func() *cobra.Command { return cmd },
					func() []string { return nil },
				),
				fx.Invoke(func(p struct {
					fx.In
					Sources srcs.Interface
					Logger  common.Logger
					LC      fx.Lifecycle
				}) {
					p.LC.Append(fx.Hook{
						OnStart: func(context.Context) error {
							params := &sources.Params{
								SourceManager: p.Sources,
								Logger:        p.Logger,
							}
							if err := sources.ExecuteList(*params); err != nil {
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

			// Test lifecycle
			app.RequireStart()
			app.RequireStop()
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
					sources: []srcs.Config{
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
					sources: []srcs.Config{},
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
					func() srcs.Interface { return sm },
					func() common.Logger { return ml },
					fx.Annotate(
						func() config.Interface { return mc },
						fx.As(new(config.Interface)),
					),
				),
				fx.Invoke(func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							params := &sources.Params{
								SourceManager: sm,
								Logger:        ml,
							}
							if err := sources.ExecuteList(*params); err != nil {
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

func TestFindByName(t *testing.T) {
	t.Parallel()
	testConfigs := []srcs.Config{
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
