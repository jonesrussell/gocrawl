package storage

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewMappingService(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := NewMockStorage()

	service := NewMappingService(mockLogger, mockStorage)
	assert.NotNil(t, service)
}

func TestGetCurrentMapping(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := NewMockStorage()
	service := NewMappingService(mockLogger, mockStorage)

	ctx := context.Background()
	expectedMapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// Reset mock expectations
	mockStorage.ExpectedCalls = nil
	mockStorage.On("GetMapping", ctx, "test-index").Return(expectedMapping, nil)

	mapping, err := service.GetCurrentMapping(ctx, "test-index")
	assert.NoError(t, err)
	assert.Equal(t, expectedMapping, mapping)
	mockStorage.AssertExpectations(t)
}

func TestUpdateMapping(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := NewMockStorage()
	service := NewMappingService(mockLogger, mockStorage)

	ctx := context.Background()
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// Reset mock expectations
	mockStorage.ExpectedCalls = nil
	mockStorage.On("UpdateMapping", ctx, "test-index", mapping).Return(nil)

	err := service.UpdateMapping(ctx, "test-index", mapping)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestValidateMapping(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := NewMockStorage()
	service := NewMappingService(mockLogger, mockStorage)

	ctx := context.Background()
	expectedMapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	tests := []struct {
		name            string
		currentMapping  map[string]interface{}
		expectedMapping map[string]interface{}
		expectedMatch   bool
	}{
		{
			name:            "mappings match",
			currentMapping:  expectedMapping,
			expectedMapping: expectedMapping,
			expectedMatch:   true,
		},
		{
			name: "mappings don't match",
			currentMapping: map[string]interface{}{
				"mappings": map[string]interface{}{
					"properties": map[string]interface{}{
						"field1": map[string]interface{}{
							"type": "text",
						},
					},
				},
			},
			expectedMapping: expectedMapping,
			expectedMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock expectations
			mockStorage.ExpectedCalls = nil
			mockStorage.On("GetMapping", ctx, "test-index").Return(tt.currentMapping, nil)

			match, err := service.ValidateMapping(ctx, "test-index", tt.expectedMapping)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedMatch, match)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestEnsureMapping(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	mockStorage := NewMockStorage()
	service := NewMappingService(mockLogger, mockStorage)

	ctx := context.Background()
	expectedMapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"field1": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	tests := []struct {
		name           string
		indexExists    bool
		currentMapping map[string]interface{}
		expectCreate   bool
		expectUpdate   bool
	}{
		{
			name:           "index doesn't exist",
			indexExists:    false,
			currentMapping: nil,
			expectCreate:   true,
			expectUpdate:   false,
		},
		{
			name:           "index exists, mapping matches",
			indexExists:    true,
			currentMapping: expectedMapping,
			expectCreate:   false,
			expectUpdate:   false,
		},
		{
			name:        "index exists, mapping doesn't match",
			indexExists: true,
			currentMapping: map[string]interface{}{
				"mappings": map[string]interface{}{
					"properties": map[string]interface{}{
						"field1": map[string]interface{}{
							"type": "text",
						},
					},
				},
			},
			expectCreate: false,
			expectUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock expectations
			mockStorage.ExpectedCalls = nil
			mockLogger.ExpectedCalls = nil

			mockStorage.On("IndexExists", ctx, "test-index").Return(tt.indexExists, nil)
			if tt.indexExists {
				mockStorage.On("GetMapping", ctx, "test-index").Return(tt.currentMapping, nil)
			}
			if tt.expectCreate {
				mockStorage.On("CreateIndex", ctx, "test-index", expectedMapping).Return(nil)
				mockLogger.On("Info", "Created new index with mapping", "index", "test-index").Return()
			}
			if tt.expectUpdate {
				mockStorage.On("UpdateMapping", ctx, "test-index", expectedMapping).Return(nil)
				mockLogger.On("Info", "Updating index mapping", "index", "test-index").Return()
				mockLogger.On("Info", "Successfully updated index mapping", "index", "test-index").Return()
			}

			err := service.EnsureMapping(ctx, "test-index", expectedMapping)
			assert.NoError(t, err)
			mockStorage.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
