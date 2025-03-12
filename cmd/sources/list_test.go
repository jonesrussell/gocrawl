// Package sources_test implements tests for the sources command package.
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// ErrSourceNotFound is returned when a source is not found
var ErrSourceNotFound = errors.New("source not found")

// mockLogger implements common.Logger for testing
type mockLogger struct {
	common.Logger
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Info(msg string, _ ...interface{}) { m.infoMessages = append(m.infoMessages, msg) }
func (m *mockLogger) Error(msg string, _ ...interface{}) {
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

func (m *mockConfig) GetString(_ string) string                  { return "" }
func (m *mockConfig) GetStringSlice(_ string) []string           { return nil }
func (m *mockConfig) GetInt(_ string) int                        { return 0 }
func (m *mockConfig) GetBool(_ string) bool                      { return false }
func (m *mockConfig) GetDuration(_ string) time.Duration         { return 0 }
func (m *mockConfig) UnmarshalKey(_ string, _ interface{}) error { return nil }
func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig    { return &config.CrawlerConfig{} }
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{}
}
func (m *mockConfig) GetLogConfig() *config.LogConfig       { return &config.LogConfig{} }
func (m *mockConfig) GetAppConfig() *config.AppConfig       { return &config.AppConfig{} }
func (m *mockConfig) GetServerConfig() *config.ServerConfig { return &config.ServerConfig{} }
func (m *mockConfig) GetSources() []config.Source           { return m.sources }

func Test_listCommand(t *testing.T) {
	t.Parallel()
	cmd := sources.ListCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List all configured content sources", cmd.Short)
	assert.Contains(t, cmd.Long, "Display a list of all content sources configured in sources.yml")
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
				sources := []srcs.Config{
					{
						Name:         "Test Source",
						URL:          "https://test.com",
						Index:        "test_content",
						ArticleIndex: "test_articles",
						RateLimit:    "1s",
						MaxDepth:     2,
					},
				}
				return cmd, &mockSourceManager{sources: sources}, &mockLogger{}, newMockConfig()
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
			app := fx.New(
				fx.NopLogger,
				fx.Supply(cmd.Context()),
				fx.Provide(
					// Provide source manager with the correct name tag
					fx.Annotate(
						func() srcs.Interface { return sm },
						fx.ResultTags(`name:"sourceManager"`),
					),
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
					Sources srcs.Interface `name:"sourceManager"`
					Logger  common.Logger
					LC      fx.Lifecycle
				}) {
					p.LC.Append(fx.Hook{
						OnStart: func(context.Context) error {
							params := &sources.ListParams{
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

			err := app.Start(cmd.Context())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, app.Stop(cmd.Context()))
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
	}{
		{
			name: "with sources",
			sources: []srcs.Config{
				{
					Name:         "Test Source",
					URL:          "https://test.com",
					Index:        "test_content",
					ArticleIndex: "test_articles",
					RateLimit:    "1s",
					MaxDepth:     2,
				},
			},
			wantErr:      false,
			wantLogCount: 0,
		},
		{
			name:         "no sources",
			sources:      nil,
			wantErr:      false,
			wantLogCount: 1, // "No sources found" message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ml := &mockLogger{}
			params := sources.ListParams{
				SourceManager: &mockSourceManager{sources: tt.sources},
				Logger:        ml,
			}

			err := sources.ExecuteList(params)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, ml.infoMessages, tt.wantLogCount)
			}
		})
	}
}

func Test_printSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sources []srcs.Config
		wantErr bool
	}{
		{
			name: "print multiple sources",
			sources: []srcs.Config{
				{
					Name:         "Test Source 1",
					URL:          "https://test1.com",
					Index:        "test1_content",
					ArticleIndex: "test1_articles",
					RateLimit:    "1s",
					MaxDepth:     2,
				},
				{
					Name:         "Test Source 2",
					URL:          "https://test2.com",
					Index:        "test2_content",
					ArticleIndex: "test2_articles",
					RateLimit:    "2s",
					MaxDepth:     3,
				},
			},
			wantErr: false,
		},
		{
			name:    "print empty sources",
			sources: []srcs.Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ml := &mockLogger{}
			err := sources.PrintSources(tt.sources, ml)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
