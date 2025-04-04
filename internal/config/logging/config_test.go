package logging

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
				Level:      "info",
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: false,
		},
		{
			name: "valid file configuration",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "file",
				File:       "test.log",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: false,
		},
		{
			name: "missing level",
			config: &Config{
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid level",
			config: &Config{
				Level:      "invalid",
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "missing encoding",
			config: &Config{
				Level:      "info",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid encoding",
			config: &Config{
				Level:      "info",
				Encoding:   "invalid",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "missing output",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid output",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "invalid",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "file output without file path",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "file",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid max size",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    -1,
				MaxBackups: 3,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid max backups",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: -1,
				MaxAge:     30,
			},
			wantErr: true,
		},
		{
			name: "invalid max age",
			config: &Config{
				Level:      "info",
				Encoding:   "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     -1,
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
				Level:      DefaultLevel,
				Encoding:   DefaultEncoding,
				Output:     DefaultOutput,
				Debug:      DefaultDebug,
				Caller:     DefaultCaller,
				Stacktrace: DefaultStacktrace,
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
				Compress:   true,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithLevel("debug"),
				WithEncoding("console"),
				WithOutput("file"),
				WithFile("custom.log"),
				WithDebug(true),
				WithCaller(true),
				WithStacktrace(true),
				WithMaxSize(200),
				WithMaxBackups(5),
				WithMaxAge(60),
				WithCompress(false),
			},
			expected: &Config{
				Level:      "debug",
				Encoding:   "console",
				Output:     "file",
				File:       "custom.log",
				Debug:      true,
				Caller:     true,
				Stacktrace: true,
				MaxSize:    200,
				MaxBackups: 5,
				MaxAge:     60,
				Compress:   false,
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
