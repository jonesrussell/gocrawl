package server

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
			name: "valid configuration with security disabled",
			config: &Config{
				SecurityEnabled: false,
				APIKey:          "",
			},
			wantErr: false,
		},
		{
			name: "valid configuration with security enabled",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          "test_id:test_api_key",
			},
			wantErr: false,
		},
		{
			name: "security enabled without API key",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          "",
			},
			wantErr: true,
		},
		{
			name: "security enabled with invalid API key format",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          "invalid_format",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key parts",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          ":",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key id",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          ":test_api_key",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key value",
			config: &Config{
				SecurityEnabled: true,
				APIKey:          "test_id:",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
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
				SecurityEnabled: DefaultSecurityEnabled,
				APIKey:          DefaultAPIKey,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithSecurityEnabled(true),
				WithAPIKey("test_id:test_api_key"),
			},
			expected: &Config{
				SecurityEnabled: true,
				APIKey:          "test_id:test_api_key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
