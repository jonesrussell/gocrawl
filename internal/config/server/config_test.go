package server_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *server.Config
		wantErr bool
	}{
		{
			name: "valid configuration with security disabled",
			config: &server.Config{
				SecurityEnabled: false,
				APIKey:          "",
			},
			wantErr: false,
		},
		{
			name: "valid configuration with security enabled",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "test_id:test_api_key",
			},
			wantErr: false,
		},
		{
			name: "security enabled without API key",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "",
			},
			wantErr: true,
		},
		{
			name: "security enabled with invalid API key format",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          "invalid_format",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key parts",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          ":",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key id",
			config: &server.Config{
				SecurityEnabled: true,
				APIKey:          ":test_api_key",
			},
			wantErr: true,
		},
		{
			name: "security enabled with empty API key value",
			config: &server.Config{
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
		opts     []server.Option
		expected *server.Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &server.Config{
				SecurityEnabled: server.DefaultSecurityEnabled,
				APIKey:          server.DefaultAPIKey,
			},
		},
		{
			name: "custom configuration",
			opts: []server.Option{
				server.WithSecurityEnabled(true),
				server.WithAPIKey("test_id:test_api_key"),
			},
			expected: &server.Config{
				SecurityEnabled: true,
				APIKey:          "test_id:test_api_key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := server.New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
