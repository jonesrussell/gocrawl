// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/indices/test"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name        string
		source      *config.Source
		indices     []string
		force       bool
		sourceName  string
		setupMocks  func(*test.MockStorage, *test.MockSources)
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				msrc.On("FindByName", "test-source").Return(nil).Once()
			},
			wantErr:     true,
			errContains: "source not found: test-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockStore := &test.MockStorage{}
			mockSources := &test.MockSources{}

			// Setup mock expectations
			tt.setupMocks(mockStore, mockSources)

			// Create test app
			app := fxtest.New(t,
				fx.NopLogger,
				test.TestModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return mockSources },
					func() []string { return tt.indices },
					func() bool { return tt.force },
					func() string { return tt.sourceName },
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
			mockSources.AssertExpectations(t)
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
		tt := tt // Capture range variable
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
