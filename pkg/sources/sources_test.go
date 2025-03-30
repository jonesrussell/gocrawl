package sources_test

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/pkg/logger"
	"github.com/jonesrussell/gocrawl/pkg/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConfig implements sources.Interface for testing
type mockConfig struct {
	sources []sources.Config
}

func (m *mockConfig) GetSources() ([]sources.Config, error) {
	return m.sources, nil
}

func (m *mockConfig) FindByName(name string) (*sources.Config, error) {
	for _, source := range m.sources {
		if source.Name == name {
			return &source, nil
		}
	}
	return nil, sources.ErrSourceNotFound
}

func (m *mockConfig) Validate(source *sources.Config) error {
	if source == nil {
		return sources.ErrInvalidSource
	}
	if source.Name == "" {
		return sources.ErrInvalidSource
	}
	if source.URL == "" {
		return sources.ErrInvalidSource
	}
	return nil
}

// TestNewSources tests the NewSources function
func TestNewSources(t *testing.T) {
	tests := []struct {
		name    string
		params  sources.Params
		wantErr bool
	}{
		{
			name: "valid params",
			params: sources.Params{
				Logger: logger.NewNoOp(),
			},
			wantErr: false,
		},
		{
			name:    "missing logger",
			params:  sources.Params{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := sources.NewSources(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, s)
		})
	}
}

// TestSources_CRUD tests the CRUD operations for sources
func TestSources_CRUD(t *testing.T) {
	params := sources.Params{
		Logger: logger.NewNoOp(),
	}
	s, err := sources.NewSources(params)
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()

	// Test adding a source
	source := &sources.Source{
		Name:     "test",
		URL:      "http://example.com",
		MaxDepth: 2,
		Time: struct {
			PublishedAt string `json:"published_at" yaml:"published_at"`
			UpdatedAt   string `json:"updated_at" yaml:"updated_at"`
		}{
			PublishedAt: "time",
			UpdatedAt:   "time",
		},
	}

	err = s.AddSource(ctx, source)
	require.NoError(t, err)

	// Test getting a source
	got, err := s.GetSource(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, source, got)

	// Test listing sources
	list, err := s.ListSources(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, source, list[0])

	// Test updating a source
	source.MaxDepth = 3
	err = s.UpdateSource(ctx, source)
	require.NoError(t, err)

	got, err = s.GetSource(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, source, got)

	// Test deleting a source
	err = s.DeleteSource(ctx, "test")
	require.NoError(t, err)

	_, err = s.GetSource(ctx, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source not found")
}

// TestSources_Validation tests the validation for sources
func TestSources_Validation(t *testing.T) {
	params := sources.Params{
		Logger: logger.NewNoOp(),
	}
	s, err := sources.NewSources(params)
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()

	tests := []struct {
		name    string
		source  *sources.Source
		wantErr bool
	}{
		{
			name:    "nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name: "empty name",
			source: &sources.Source{
				URL:      "http://example.com",
				MaxDepth: 2,
			},
			wantErr: true,
		},
		{
			name: "empty URL",
			source: &sources.Source{
				Name:     "test",
				MaxDepth: 2,
			},
			wantErr: true,
		},
		{
			name: "negative max depth",
			source: &sources.Source{
				Name:     "test",
				URL:      "http://example.com",
				MaxDepth: -1,
			},
			wantErr: true,
		},
		{
			name: "valid source",
			source: &sources.Source{
				Name:     "test",
				URL:      "http://example.com",
				MaxDepth: 2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.AddSource(ctx, tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestSources_Metrics tests the metrics for sources
func TestSources_Metrics(t *testing.T) {
	params := sources.Params{
		Logger: logger.NewNoOp(),
	}
	s, err := sources.NewSources(params)
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()

	// Initial metrics
	metrics := s.GetMetrics()
	assert.Equal(t, int64(0), metrics.SourceCount)
	assert.NotZero(t, metrics.LastUpdated)

	// Add a source
	source := &sources.Source{
		Name:     "test",
		URL:      "http://example.com",
		MaxDepth: 2,
	}

	err = s.AddSource(ctx, source)
	require.NoError(t, err)

	// Check updated metrics
	metrics = s.GetMetrics()
	assert.Equal(t, int64(1), metrics.SourceCount)
	assert.NotZero(t, metrics.LastUpdated)

	// Delete the source
	err = s.DeleteSource(ctx, "test")
	require.NoError(t, err)

	// Check final metrics
	metrics = s.GetMetrics()
	assert.Equal(t, int64(0), metrics.SourceCount)
	assert.NotZero(t, metrics.LastUpdated)
}

// TestValidateConfig tests the ValidateConfig function
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *sources.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &sources.Config{
				Name:      "test",
				URL:       "http://test.com",
				RateLimit: time.Second,
				MaxDepth:  2,
				Time:      []string{"00:00"},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing name",
			config: &sources.Config{
				URL:       "http://test.com",
				RateLimit: time.Second,
				MaxDepth:  2,
				Time:      []string{"00:00"},
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			config: &sources.Config{
				Name:      "test",
				RateLimit: time.Second,
				MaxDepth:  2,
				Time:      []string{"00:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: &sources.Config{
				Name:      "test",
				URL:       "not-a-url",
				RateLimit: time.Second,
				MaxDepth:  2,
				Time:      []string{"00:00"},
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			config: &sources.Config{
				Name:      "test",
				URL:       "http://example.com",
				RateLimit: -time.Second,
				MaxDepth:  2,
				Time:      []string{"00:00"},
			},
			wantErr: true,
		},
		{
			name: "negative max depth",
			config: &sources.Config{
				Name:      "test",
				URL:       "http://example.com",
				RateLimit: time.Second,
				MaxDepth:  -2,
				Time:      []string{"00:00"},
			},
			wantErr: true,
		},
		{
			name: "invalid time format",
			config: &sources.Config{
				Name:      "test",
				URL:       "http://example.com",
				RateLimit: time.Second,
				MaxDepth:  2,
				Time:      []string{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sources.ValidateConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
