package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				Environment: "development",
				Name:        "test",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "missing environment",
			config: &Config{
				Name:    "test",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: &Config{
				Environment: "invalid",
				Name:        "test",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			config: &Config{
				Environment: "development",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			config: &Config{
				Environment: "development",
				Name:        "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     []Option
		expected *Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &Config{
				Environment: "development",
				Name:        "gocrawl",
				Version:     "0.1.0",
				Debug:       false,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithEnvironment("production"),
				WithName("custom"),
				WithVersion("2.0.0"),
				WithDebug(true),
			},
			expected: &Config{
				Environment: "production",
				Name:        "custom",
				Version:     "2.0.0",
				Debug:       true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
