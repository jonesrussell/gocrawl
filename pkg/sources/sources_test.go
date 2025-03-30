package sources_test

import (
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
				Config: &mockConfig{
					sources: []sources.Config{
						{
							Name:      "test",
							URL:       "http://test.com",
							RateLimit: time.Second,
							MaxDepth:  2,
							Time:      []string{"00:00"},
						},
					},
				},
				Logger: logger.NewNoOp(),
			},
			wantErr: false,
		},
		{
			name: "missing config",
			params: sources.Params{
				Logger: logger.NewNoOp(),
			},
			wantErr: true,
		},
		{
			name: "missing logger",
			params: sources.Params{
				Config: &mockConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sources, err := sources.NewSources(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, sources)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, sources)
		})
	}
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
