// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			// Create mock storage
			mockStore := &mockStorage{}

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return(tt.indices, tt.listError)
			if tt.listError == nil && len(tt.indices) > 0 {
				for _, index := range tt.indices {
					mockStore.On("DeleteIndex", mock.Anything, index).Return(tt.deleteError)
				}
			}

			// Create test app
			app := fxtest.New(t,
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return nil },
					func() []string { return tt.indices },
					func() bool { return true }, // Auto-confirm deletion
					indices.NewDeleter,
				),
				fx.Invoke(func(lc fx.Lifecycle, deleter *indices.Deleter, ctx context.Context) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := deleter.Start(ctx)
							if tt.listError != nil || tt.deleteError != nil {
								require.Error(t, err)
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
		})
	}
}
