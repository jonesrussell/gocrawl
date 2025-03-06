package storage

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// MappingServiceInterface defines the interface for mapping management
type MappingServiceInterface interface {
	// EnsureMapping ensures the index mapping matches the expected mapping
	EnsureMapping(ctx context.Context, index string, expectedMapping map[string]interface{}) error
	// GetCurrentMapping gets the current mapping for an index
	GetCurrentMapping(ctx context.Context, index string) (map[string]interface{}, error)
	// UpdateMapping updates the mapping for an index
	UpdateMapping(ctx context.Context, index string, mapping map[string]interface{}) error
	// ValidateMapping validates if the current mapping matches the expected mapping
	ValidateMapping(ctx context.Context, index string, expectedMapping map[string]interface{}) (bool, error)
}

// MappingService implements the MappingServiceInterface
type MappingService struct {
	logger  logger.Interface
	storage Interface
}

// NewMappingService creates a new MappingService instance
func NewMappingService(logger logger.Interface, storage Interface) MappingServiceInterface {
	return &MappingService{
		logger:  logger,
		storage: storage,
	}
}

// GetCurrentMapping gets the current mapping for an index
func (s *MappingService) GetCurrentMapping(ctx context.Context, index string) (map[string]interface{}, error) {
	return s.storage.GetMapping(ctx, index)
}

// UpdateMapping updates the mapping for an index
func (s *MappingService) UpdateMapping(ctx context.Context, index string, mapping map[string]interface{}) error {
	return s.storage.UpdateMapping(ctx, index, mapping)
}

// ValidateMapping validates if the current mapping matches the expected mapping
func (s *MappingService) ValidateMapping(
	ctx context.Context,
	index string,
	expectedMapping map[string]interface{},
) (bool, error) {
	currentMapping, err := s.GetCurrentMapping(ctx, index)
	if err != nil {
		return false, fmt.Errorf("failed to get current mapping: %w", err)
	}

	// Compare the mappings
	return reflect.DeepEqual(currentMapping, expectedMapping), nil
}

// EnsureMapping ensures the index mapping matches the expected mapping
func (s *MappingService) EnsureMapping(
	ctx context.Context,
	index string,
	expectedMapping map[string]interface{},
) error {
	exists, err := s.storage.IndexExists(ctx, index)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if !exists {
		s.logger.Info("Creating new index with mapping", "index", index)
		if createErr := s.storage.CreateIndex(ctx, index, expectedMapping); createErr != nil {
			return fmt.Errorf("failed to create index: %w", createErr)
		}
		return nil
	}

	match, err := s.ValidateMapping(ctx, index, expectedMapping)
	if err != nil {
		return fmt.Errorf("failed to validate mapping: %w", err)
	}

	if !match {
		s.logger.Info("Updating index mapping", "index", index)
		if updateErr := s.UpdateMapping(ctx, index, expectedMapping); updateErr != nil {
			return fmt.Errorf("failed to update mapping: %w", updateErr)
		}
		s.logger.Info("Successfully updated index mapping", "index", index)
	}

	return nil
}
