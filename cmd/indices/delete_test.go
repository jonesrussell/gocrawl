// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/indices/test"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name        string
		indices     []string
		listError   error
		deleteError error
	}{
		{
			name:        "delete existing indices",
			indices:     []string{"test-index-1", "test-index-2"},
			listError:   nil,
			deleteError: nil,
		},
		{
			name:        "no indices to delete",
			indices:     []string{},
			listError:   nil,
			deleteError: nil,
		},
		{
			name:        "list indices error",
			indices:     []string{},
			listError:   assert.AnError,
			deleteError: nil,
		},
		{
			name:        "delete error",
			indices:     []string{"test-index"},
			listError:   nil,
			deleteError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return(tt.indices, tt.listError)
			if tt.listError == nil {
				for _, index := range tt.indices {
					mockStore.On("DeleteIndex", mock.Anything, index).Return(tt.deleteError)
				}
			}

			// Create test app
			app := fx.New(
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return nil },
					func() []string { return tt.indices },
					func() bool { return false },
					indices.NewDeleter,
				),
				fx.Invoke(func(deleter *indices.Deleter) {
					err := deleter.Start(context.Background())
					if tt.listError != nil || tt.deleteError != nil {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
					}
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
