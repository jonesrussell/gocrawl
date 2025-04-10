package app_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *app.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &app.Config{
				Environment: "development",
				Name:        "test",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "missing environment",
			config: &app.Config{
				Name:    "test",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: &app.Config{
				Environment: "invalid",
				Name:        "test",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			config: &app.Config{
				Environment: "development",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			config: &app.Config{
				Environment: "development",
				Name:        "test",
			},
			wantErr: true,
		},
	}

	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.config.Validate()
			if test.wantErr {
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
		opts     []app.Option
		expected *app.Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &app.Config{
				Environment: "development",
				Name:        "gocrawl",
				Version:     "0.1.0",
				Debug:       false,
			},
		},
		{
			name: "custom configuration",
			opts: []app.Option{
				app.WithEnvironment("production"),
				app.WithName("custom"),
				app.WithVersion("2.0.0"),
				app.WithDebug(true),
			},
			expected: &app.Config{
				Environment: "production",
				Name:        "custom",
				Version:     "2.0.0",
				Debug:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := app.New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
