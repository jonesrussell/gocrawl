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
	"go.uber.org/fx/fxtest"
)

func TestListCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		indices   []string
		health    string
		docCount  int64
		listError error
	}{
		{
			name:      "list existing indices",
			indices:   []string{"test-index-1", "test-index-2"},
			health:    "green",
			docCount:  100,
			listError: nil,
		},
		{
			name:      "no indices",
			indices:   []string{},
			health:    "",
			docCount:  0,
			listError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return(tt.indices, tt.listError)
			if tt.listError == nil && len(tt.indices) > 0 {
				for _, index := range tt.indices {
					mockStore.On("GetIndexHealth", mock.Anything, index).Return(tt.health, nil)
					mockStore.On("GetIndexDocCount", mock.Anything, index).Return(tt.docCount, nil)
				}
			}

			// Create test app
			app := fxtest.New(t,
				fx.NopLogger,
				test.TestModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return nil },
					indices.NewLister,
				),
				fx.Invoke(func(lc fx.Lifecycle, lister *indices.Lister, ctx context.Context) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := lister.Start(ctx)
							if tt.listError != nil {
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

// TestListCommandErrors tests error handling in the list command
func TestListCommandErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		indices     []string
		healthError error
		docError    error
	}{
		{
			name:        "health check error",
			indices:     []string{"test-index"},
			healthError: assert.AnError,
			docError:    nil,
		},
		{
			name:        "doc count error",
			indices:     []string{"test-index"},
			healthError: nil,
			docError:    assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("TestConnection", mock.Anything).Return(nil)
			mockStore.On("ListIndices", mock.Anything).Return(tt.indices, nil)
			for _, index := range tt.indices {
				if tt.healthError != nil {
					mockStore.On("GetIndexHealth", mock.Anything, index).Return("", tt.healthError)
				} else {
					mockStore.On("GetIndexHealth", mock.Anything, index).Return("green", nil)
					mockStore.On("GetIndexDocCount", mock.Anything, index).Return(int64(0), tt.docError)
				}
			}

			// Create test app
			app := fxtest.New(t,
				fx.NopLogger,
				test.TestModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					func() sources.Interface { return nil },
					indices.NewLister,
				),
				fx.Invoke(func(lc fx.Lifecycle, lister *indices.Lister, ctx context.Context) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := lister.Start(ctx)
							require.Error(t, err)
							if tt.healthError != nil {
								require.ErrorIs(t, err, tt.healthError)
							} else {
								require.ErrorIs(t, err, tt.docError)
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
