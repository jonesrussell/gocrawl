// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// MockSources implements sources.Interface for testing
type MockSources struct {
	mock.Mock
}

func (m *MockSources) FindByName(name string) *sourceutils.SourceConfig {
	args := m.Called(name)
	if source, ok := args.Get(0).(*sourceutils.SourceConfig); ok {
		return source
	}
	return nil
}

func (m *MockSources) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) DeleteSource(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockSources) GetMetrics() sources.Metrics {
	args := m.Called()
	if metrics, ok := args.Get(0).(sources.Metrics); ok {
		return metrics
	}
	return sources.Metrics{}
}

func (m *MockSources) GetSources() []sources.Source {
	args := m.Called()
	if sources, ok := args.Get(0).([]sources.Source); ok {
		return sources
	}
	return nil
}

func TestDeleteCommand(t *testing.T) {
	// Set up test environment
	cleanup := configtestutils.SetupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name        string
		source      *config.Source
		indices     []string
		force       bool
		sourceName  string
		setupMocks  func(*testutils.MockStorage, *MockSources)
		wantErr     bool
		errContains string
	}{
		{
			name: "successfully deletes index",
			source: &config.Source{
				Name:  "test-source",
				Index: "test-index",
			},
			indices:    []string{"test-index"},
			force:      true,
			sourceName: "",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successfully deletes source indices",
			source: &config.Source{
				Name:         "test-source",
				Index:        "test-index",
				ArticleIndex: "test-articles",
			},
			indices:    []string{},
			force:      true,
			sourceName: "test-source",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index", "test-articles"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
				ms.On("DeleteIndex", mock.Anything, "test-articles").Return(nil)
				msrc.On("FindByName", "test-source").Return(&sourceutils.SourceConfig{
					Name:         "test-source",
					Index:        "test-index",
					ArticleIndex: "test-articles",
				}).Once()
			},
			wantErr: false,
		},
		{
			name: "successfully deletes source indices with spaces",
			source: &config.Source{
				Name:         "test source",
				Index:        "test-index",
				ArticleIndex: "test-articles",
			},
			indices:    []string{},
			force:      true,
			sourceName: "test source",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index", "test-articles"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
				ms.On("DeleteIndex", mock.Anything, "test-articles").Return(nil)
				msrc.On("FindByName", "test source").Return(&sourceutils.SourceConfig{
					Name:         "test source",
					Index:        "test-index",
					ArticleIndex: "test-articles",
				}).Once()
			},
			wantErr: false,
		},
		{
			name: "source not found",
			source: &config.Source{
				Name: "nonexistent",
			},
			indices:    []string{},
			force:      true,
			sourceName: "nonexistent",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				msrc.On("FindByName", "nonexistent").Return(nil).Once()
			},
			wantErr:     true,
			errContains: "source not found",
		},
		{
			name: "connection test fails",
			source: &config.Source{
				Name:  "test-source",
				Index: "test-index",
			},
			indices:    []string{"test-index"},
			force:      true,
			sourceName: "",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(assert.AnError)
			},
			wantErr:     true,
			errContains: "failed to connect to storage",
		},
		{
			name: "list indices fails",
			source: &config.Source{
				Name:  "test-source",
				Index: "test-index",
			},
			indices:    []string{"test-index"},
			force:      true,
			sourceName: "",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return(nil, assert.AnError)
			},
			wantErr:     true,
			errContains: "assert.AnError general error for testing",
		},
		{
			name: "delete fails",
			source: &config.Source{
				Name:  "test-source",
				Index: "test-index",
			},
			indices:    []string{"test-index"},
			force:      true,
			sourceName: "",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:    "no args and no source",
			source:  nil,
			indices: []string{},
			force:   false,
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{}, nil)
			},
			wantErr:     true,
			errContains: "no indices specified",
		},
		{
			name:       "args and source",
			source:     nil,
			indices:    []string{"test-index"},
			force:      false,
			sourceName: "test-source",
			setupMocks: func(ms *testutils.MockStorage, msrc *MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				msrc.On("FindByName", "test-source").Return(nil).Once()
			},
			wantErr:     true,
			errContains: "source not found: test-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupMocks := func() (*testutils.MockStorage, *MockSources) {
				mockStore := &testutils.MockStorage{}
				mockSources := &MockSources{}

				if tt.setupMocks != nil {
					tt.setupMocks(mockStore, mockSources)
				}

				return mockStore, mockSources
			}

			mockStore, mockSources := setupMocks()

			// Create test app
			app := fx.New(
				fx.NopLogger,
				logger.Module,
				fx.Provide(
					fx.Annotate(
						func() *testing.T { return t },
						fx.ResultTags(`name:"test"`),
					),
					fx.Annotate(
						configtestutils.NewTestLogger,
						fx.ParamTags(`name:"test"`),
					),
					func() logger.Config {
						return logger.Config{
							Level:       logger.DebugLevel,
							Development: true,
							Encoding:    "console",
						}
					},
					func() logger.Params {
						return logger.Params{
							Config: &logger.Config{
								Level:       logger.DebugLevel,
								Development: true,
								Encoding:    "console",
							},
						}
					},
					func() config.Interface {
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
						mockConfig.On("GetCrawlerConfig").Return(&config.CrawlerConfig{
							BaseURL:     "http://localhost",
							MaxDepth:    2,
							RateLimit:   time.Second * 2,
							RandomDelay: time.Second,
							IndexName:   "test-index",
							SourceFile:  "testdata/sources.yml",
							Parallelism: 2,
						})
						mockConfig.On("Validate").Return(nil)
						return mockConfig
					},
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return mockSources },
					func() []string { return tt.indices },
					func() bool { return tt.force },
					func() string { return tt.sourceName },
					func() context.Context { return t.Context() },
					indices.NewDeleter,
				),
				fx.Invoke(func(lc fx.Lifecycle, deleter *indices.Deleter, ctx context.Context) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := deleter.Start(ctx)
							if tt.wantErr {
								require.Error(t, err)
								if tt.errContains != "" {
									require.Contains(t, err.Error(), tt.errContains)
								}
								return err // Propagate the error when we expect it
							}
							require.NoError(t, err)
							return nil
						},
						OnStop: func(ctx context.Context) error {
							return nil
						},
					})
				}),
			)

			if err := app.Start(context.Background()); err != nil {
				if tt.wantErr {
					require.Error(t, err)
					if tt.errContains != "" {
						require.Contains(t, err.Error(), tt.errContains)
					}
					return
				}
				require.NoError(t, err)
			}

			// Ensure app is stopped properly
			defer func() {
				if err := app.Stop(context.Background()); err != nil {
					t.Logf("Error stopping application: %v", err)
				}
			}()
		})
	}
}

func TestDeleteCommandArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		sourceName string
		wantErr    bool
		errMsg     string
	}{
		{
			name:    "no args and no source",
			args:    []string{},
			wantErr: true,
			errMsg:  "either specify indices or use --source flag",
		},
		{
			name:       "args with source",
			args:       []string{"index1"},
			sourceName: "test-source",
			wantErr:    true,
			errMsg:     "cannot specify both indices and --source flag",
		},
		{
			name:       "source with spaces",
			sourceName: "test source",
			args:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call ValidateDeleteArgs directly instead of executing the command
			err := indices.ValidateDeleteArgs(tt.sourceName, tt.args)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func setupTestDeps() *Dependencies {
	return &Dependencies{
		Config: &mockConfig{
			app: &config.App{
				Elasticsearch: &elasticsearch.Config{
					URL: "http://localhost:9200",
				},
				Log: &log.Config{
					Level: "debug",
				},
				Priority: &priority.Config{
					Enabled: true,
				},
				Server: &server.Config{
					Port: 8080,
				},
			},
		},
		Sources: []*sourceutils.SourceConfig{
			{
				Name:         "test-source",
				URL:          "https://example.com",
				ArticleIndex: "test-articles",
				Index:        "test-content",
			},
		},
	}
}
