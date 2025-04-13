package logging_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config/logging"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *logging.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
			config: &logging.Config{
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
		opts     []logging.Option
		expected *logging.Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &logging.Config{
				Level:      logging.DefaultLevel,
				Encoding:   logging.DefaultEncoding,
				Output:     logging.DefaultOutput,
				Debug:      logging.DefaultDebug,
				Caller:     logging.DefaultCaller,
				Stacktrace: logging.DefaultStacktrace,
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
				Compress:   true,
			},
		},
		{
			name: "custom configuration",
			opts: []logging.Option{
				logging.WithLevel("debug"),
				logging.WithEncoding("console"),
				logging.WithOutput("file"),
				logging.WithFile("custom.log"),
				logging.WithDebug(true),
				logging.WithCaller(true),
				logging.WithStacktrace(true),
				logging.WithMaxSize(200),
				logging.WithMaxBackups(5),
				logging.WithMaxAge(60),
				logging.WithCompress(false),
			},
			expected: &logging.Config{
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := logging.New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
