// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/indices/test"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		source      *config.Source
		indices     []string
		force       bool
		sourceName  string
		setupMocks  func(*test.MockStorage, *test.MockSources, *test.MockConfig)
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
				mc.On("GetCommand").Return("delete")
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index", "test-articles"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
				ms.On("DeleteIndex", mock.Anything, "test-articles").Return(nil)
				msrc.On("FindByName", "test-source").Return(&sourceutils.SourceConfig{
					Name:         "test-source",
					Index:        "test-index",
					ArticleIndex: "test-articles",
				})
				mc.On("GetCommand").Return("delete")
			},
			wantErr: false,
		},
		{
			name: "successfully deletes source indices with quotes",
			source: &config.Source{
				Name:         "test source",
				Index:        "test-index",
				ArticleIndex: "test-articles",
			},
			indices:    []string{},
			force:      true,
			sourceName: `"test source"`,
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index", "test-articles"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(nil)
				ms.On("DeleteIndex", mock.Anything, "test-articles").Return(nil)
				msrc.On("FindByName", "test source").Return(&sourceutils.SourceConfig{
					Name:         "test source",
					Index:        "test-index",
					ArticleIndex: "test-articles",
				})
				mc.On("GetCommand").Return("delete")
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				msrc.On("FindByName", "nonexistent").Return(nil)
				mc.On("GetCommand").Return("delete")
			},
			wantErr:     true,
			errContains: "source not found",
		},
		{
			name: "source not found with quotes",
			source: &config.Source{
				Name: "nonexistent source",
			},
			indices:    []string{},
			force:      true,
			sourceName: `"nonexistent source"`,
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				msrc.On("FindByName", "nonexistent source").Return(nil)
				mc.On("GetCommand").Return("delete")
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(assert.AnError)
				mc.On("GetCommand").Return("delete")
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return(nil, assert.AnError)
				mc.On("GetCommand").Return("delete")
			},
			wantErr: true,
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
			setupMocks: func(ms *test.MockStorage, msrc *test.MockSources, mc *test.MockConfig) {
				ms.On("TestConnection", mock.Anything).Return(nil)
				ms.On("ListIndices", mock.Anything).Return([]string{"test-index"}, nil)
				ms.On("DeleteIndex", mock.Anything, "test-index").Return(assert.AnError)
				mc.On("GetCommand").Return("delete")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := &test.MockStorage{}
			mockSources := &test.MockSources{}
			mockConfig := &test.MockConfig{}
			mockLogger := logger.NewNoOp()

			tt.setupMocks(mockStorage, mockSources, mockConfig)

			// Set the source name for this test
			indices.DeleteSourceName = tt.sourceName

			deleter := indices.NewDeleter(mockConfig, mockLogger, mockStorage, mockSources, tt.indices, tt.force)
			err := deleter.Start(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
			mockSources.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
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
			name:       "source with quotes",
			sourceName: `"test source"`,
			args:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := indices.Command()
			args := append([]string{"delete"}, tt.args...)
			if tt.sourceName != "" {
				args = append(args, "--source", tt.sourceName)
			}
			cmd.SetArgs(args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}
