// Package indices_test provides tests for the indices command.
package indices_test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/indices/test"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// testModule provides a test module with mock dependencies
var testModule = func(t *testing.T) fx.Option {
	return fx.Module("test",
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface { return configtestutils.NewMockConfig() },
			logger.NewNoOp,
			indices.NewCreator,
		),
	)
}

func TestCreateCommand(t *testing.T) {
	tests := []struct {
		name      string
		index     string
		mapping   map[string]any
		exists    bool
		existsErr error
		createErr error
	}{
		{
			name:  "create new index",
			index: "test-index",
			mapping: map[string]any{
				"mappings": map[string]any{
					"properties": map[string]any{
						"title": map[string]any{
							"type": "text",
						},
					},
				},
			},
			exists:    false,
			existsErr: nil,
			createErr: nil,
		},
		{
			name:  "index already exists",
			index: "existing-index",
			mapping: map[string]any{
				"mappings": map[string]any{
					"properties": map[string]any{
						"title": map[string]any{
							"type": "text",
						},
					},
				},
			},
			exists:    true,
			existsErr: nil,
			createErr: nil,
		},
		{
			name:  "check existence error",
			index: "error-index",
			mapping: map[string]any{
				"mappings": map[string]any{
					"properties": map[string]any{
						"title": map[string]any{
							"type": "text",
						},
					},
				},
			},
			exists:    false,
			existsErr: assert.AnError,
			createErr: nil,
		},
		{
			name:  "create error",
			index: "create-error-index",
			mapping: map[string]any{
				"mappings": map[string]any{
					"properties": map[string]any{
						"title": map[string]any{
							"type": "text",
						},
					},
				},
			},
			exists:    false,
			existsErr: nil,
			createErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock storage
			mockStore := new(test.MockStorage)

			// Setup expectations
			mockStore.On("IndexExists", mock.Anything, tt.index).Return(tt.exists, tt.existsErr)
			if !tt.exists && tt.existsErr == nil {
				mockStore.On("CreateIndex", mock.Anything, tt.index, tt.mapping).Return(tt.createErr)
			}

			// Create test app
			app := fx.New(
				fx.NopLogger,
				testModule(t),
				fx.Provide(
					func() storagetypes.Interface { return mockStore },
					indices.NewCreator,
				),
				fx.Invoke(func(creator *indices.Creator) {
					err := creator.Start(context.Background())
					if tt.existsErr != nil || tt.createErr != nil {
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
