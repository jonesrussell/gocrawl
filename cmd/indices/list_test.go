// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/indices/test"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name           string
		indices        []string
		healthStatus   map[string]string
		docCounts      map[string]int64
		expectedOutput []string
	}{
		{
			name:    "list all non-internal indices",
			indices: []string{"content", "articles", ".kibana", ".security"},
			healthStatus: map[string]string{
				"content":   "yellow",
				"articles":  "green",
				".kibana":   "green",
				".security": "green",
			},
			docCounts: map[string]int64{
				"content":   3695,
				"articles":  6696,
				".kibana":   1,
				".security": 1,
			},
			expectedOutput: []string{
				"content",
				"articles",
			},
		},
		{
			name:           "no indices",
			indices:        []string{},
			healthStatus:   map[string]string{},
			docCounts:      map[string]int64{},
			expectedOutput: []string{},
		},
		{
			name:    "only internal indices",
			indices: []string{".kibana", ".security", ".system"},
			healthStatus: map[string]string{
				".kibana":   "green",
				".security": "green",
				".system":   "green",
			},
			docCounts: map[string]int64{
				".kibana":   1,
				".security": 1,
				".system":   1,
			},
			expectedOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return(tt.indices, nil)
			for _, index := range tt.indices {
				mockStore.On("GetIndexHealth", mock.Anything, index).Return(tt.healthStatus[index], nil)
				mockStore.On("GetIndexDocCount", mock.Anything, index).Return(tt.docCounts[index], nil)
			}

			// Create test app
			app := fx.New(
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					indices.NewLister,
				),
				fx.Invoke(func(lister *indices.Lister) {
					err := lister.Start(context.Background())
					require.NoError(t, err)
				}),
			)

			// Start the app
			err := app.Start(context.Background())
			require.NoError(t, err)

			// Verify all expected calls were made
			mockStore.AssertExpectations(t)

			// Stop the app
			err = app.Stop(context.Background())
			require.NoError(t, err)
		})
	}
}

func TestListCommandErrors(t *testing.T) {
	tests := []struct {
		name        string
		listError   error
		healthError error
		docError    error
	}{
		{
			name:        "list indices error",
			listError:   assert.AnError,
			healthError: nil,
			docError:    nil,
		},
		{
			name:        "get health error",
			listError:   nil,
			healthError: assert.AnError,
			docError:    nil,
		},
		{
			name:        "get doc count error",
			listError:   nil,
			healthError: nil,
			docError:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return([]string{"test"}, tt.listError)
			if tt.listError == nil {
				mockStore.On("GetIndexHealth", mock.Anything, "test").Return("green", tt.healthError)
				if tt.healthError == nil {
					mockStore.On("GetIndexDocCount", mock.Anything, "test").Return(int64(1), tt.docError)
				}
			}

			// Create test app
			app := fx.New(
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					indices.NewLister,
				),
				fx.Invoke(func(lister *indices.Lister) {
					err := lister.Start(context.Background())
					require.Error(t, err)
				}),
			)

			// Start the app
			err := app.Start(context.Background())
			require.NoError(t, err)

			// Verify all expected calls were made
			mockStore.AssertExpectations(t)

			// Stop the app
			err = app.Stop(context.Background())
			require.NoError(t, err)
		})
	}
}
