// Package crawler_test provides test utilities for the crawler package.
package crawler_test

import (
	"context"
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Sentinel errors for testing
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoSources      = errors.New("no sources available")
	ErrNoResults      = errors.New("no results found")
	ErrNoAggregation  = errors.New("aggregation not implemented")
)

// Mock type declarations
type (
	// mockSearchManager implements api.SearchManager for testing.
	mockSearchManager struct {
		api.SearchManager
	}

	// mockIndexManager implements api.IndexManager for testing.
	mockIndexManager struct {
		api.IndexManager
	}

	// mockStorage implements types.Interface for testing.
	mockStorage struct {
		types.Interface
	}

	// mockSources implements sources.Interface for testing.
	mockSources struct {
		sources.Interface
	}

	// mockContentProcessor implements collector.Processor for testing
	mockContentProcessor struct{}
)

// Mock implementations for other interfaces
func (m *mockSearchManager) Search(_ context.Context, _ string, _ any) ([]any, error) {
	return nil, ErrNoResults
}

func (m *mockSearchManager) Count(_ context.Context, _ string, _ any) (int64, error) {
	return 0, nil
}

func (m *mockSearchManager) Aggregate(_ context.Context, _ string, _ any) (any, error) {
	return nil, ErrNoAggregation
}

func (m *mockIndexManager) Index(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockIndexManager) Close() error {
	return nil
}

func (m *mockStorage) Store(_ context.Context, _ string, _ any) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockSources) GetSource(_ string) (*sources.Config, error) {
	return nil, ErrNotImplemented
}

func (m *mockSources) ListSources() ([]*sources.Config, error) {
	return nil, ErrNoSources
}

func (m *mockContentProcessor) Process(e *colly.HTMLElement) error {
	// No-op for testing
	return nil
}
