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
	"github.com/jonesrussell/gocrawl/internal/sources/testutils"
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
	common.Logger
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Info(msg string, _ ...any) { m.infoMessages = append(m.infoMessages, msg) }
func (m *mockLogger) Error(msg string, _ ...any) {
	m.errorMessages = append(m.errorMessages, msg)
}

// mockSourceManager implements sources.Interface for testing
type mockSourceManager struct {
	sources []srcs.Config
}

func (m *mockSourceManager) GetSources() []srcs.Config                 { return m.sources }
func (m *mockSourceManager) FindByName(_ string) (*srcs.Config, error) { return nil, ErrSourceNotFound }
func (m *mockSourceManager) Validate(_ *srcs.Config) error             { return nil }

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
	sources []config.Source
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		sources: []config.Source{
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

// testParams holds the parameters required for listing sources.
type testParams struct {
	ctx           context.Context
	sources       srcs.Interface
	logger        common.Logger
	outputFormat  string
	showMetadata  bool
	showSelectors bool
}

// setupTestApp creates a new test application with all required dependencies.
func setupTestApp(t *testing.T) *fxtest.App {
	t.Helper()

	// Create mock logger
	mockLogger := &testutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	return fxtest.New(t,
		fx.NopLogger,
		// Provide core dependencies
		fx.Provide(
			// Named dependencies
			func() srcs.Interface { return &mockSources{} },
			// Logger provider that replaces the default logger.Module provider
			fx.Annotate(
				func() common.Logger { return mockLogger },
				fx.ResultTags(`name:"logger"`),
			),
		),
		// Include test config module and sources module
		TestConfigModule,
		TestCommonModule,
		sources.Module,
		// Verify dependencies
		fx.Invoke(func(p TestParams) {
			verifyDependencies(t, &p)
		}),
	)
}

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
				return cmd, &mockSourceManager{sources: sourceConfigs}, &mockLogger{}, newMockConfig()
			},
			wantErr: false,
		},
		{
			name: "no sources found",
			setup: func(t *testing.T) (*cobra.Command, *mockSourceManager, *mockLogger, *mockConfig) {
				cmd := &cobra.Command{}
				cmd.SetContext(t.Context())
				return cmd, &mockSourceManager{sources: nil}, &mockLogger{}, newMockConfig()
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
	t.Parallel()
	tests := []struct {
		name         string
		sources      []srcs.Config
		wantErr      bool
		wantLogCount int
		wantLogMsg   string
	}{
		{
			name: "with sources",
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
			wantErr:      false,
			wantLogCount: 0,
			wantLogMsg:   "",
		},
		{
			name:         "no sources",
			sources:      nil,
			wantErr:      false,
			wantLogCount: 1,
			wantLogMsg:   "No sources found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create mock source manager with test sources
			sm := &mockSourceManager{sources: tt.sources}
			ml := &mockLogger{}

			// Create test app with mock dependencies
			app := fxtest.New(t,
				fx.Supply(fx.NopLogger), // Suppress Fx logs
				fx.Provide(
					// Provide source manager without name tag
					func() srcs.Interface { return sm },
					// Provide mock logger with the correct interface
					fx.Annotate(
						func() common.Logger { return ml },
						fx.As(new(common.Logger)),
					),
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

			// Check log messages
			require.Len(t, ml.infoMessages, tt.wantLogCount)
			if tt.wantLogCount > 0 {
				require.Contains(t, ml.infoMessages, tt.wantLogMsg)
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
	s := testutils.NewTestSources(testConfigs)
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
