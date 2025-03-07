package logger_test

import (
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewCustomLogger(t *testing.T) {
	tests := []struct {
		name    string
		params  logger.Params
		wantErr bool
	}{
		{
			name: "development logger with debug",
			params: logger.Params{
				Debug:  true,
				Level:  zapcore.DebugLevel,
				AppEnv: "development",
			},
			wantErr: false,
		},
		{
			name: "production logger",
			params: logger.Params{
				Debug:  false,
				Level:  zapcore.InfoLevel,
				AppEnv: "production",
			},
			wantErr: false,
		},
		{
			name: "default environment",
			params: logger.Params{
				Debug: true,
				Level: zapcore.DebugLevel,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel := "info"                                   // Set the desired log level
			logger, err := logger.NewDevelopmentLogger(logLevel) // Pass log level as string
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDevelopmentLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger == nil {
				t.Error("NewDevelopmentLogger() returned nil logger")
				return
			}

			// Test logging methods
			logger.Info("test info message", "key", "value")
			logger.Error("test error message", "key", "value")
			logger.Debug("test debug message", "key", "value")
			logger.Warn("test warn message", "key", "value")
			logger.Errorf("test error format %s", "value")

			// Test GetZapLogger
			if logger.GetZapLogger() == nil {
				t.Error("GetZapLogger() returned nil")
			}

			// Test Sync
			if syncErr := logger.Sync(); syncErr != nil {
				// Ignore sync errors as they're expected when writing to console
				t.Log("Sync() error (expected):", syncErr)
			}
		})
	}
}

func TestNewDevelopmentLogger(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		expectErr bool
	}{
		{
			name:      "debug level",
			levelStr:  "debug",
			expectErr: false,
		},
		{
			name:      "info level",
			levelStr:  "info",
			expectErr: false,
		},
		{
			name:      "warn level",
			levelStr:  "warn",
			expectErr: false,
		},
		{
			name:      "error level",
			levelStr:  "error",
			expectErr: false,
		},
		{
			name:      "invalid level",
			levelStr:  "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.NewDevelopmentLogger(tt.levelStr)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, log)
			}
		})
	}
}

func TestNewProductionLogger(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		expectErr bool
	}{
		{
			name:      "info level",
			levelStr:  "info",
			expectErr: false,
		},
		{
			name:      "error level",
			levelStr:  "error",
			expectErr: false,
		},
		{
			name:      "invalid level",
			levelStr:  "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.NewProductionLogger(tt.levelStr)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, log)
			}
		})
	}
}

func TestInitializeLogger(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "development environment",
			cfg: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{
					Level: "debug",
					Debug: true,
				},
			},
			wantErr: false,
		},
		{
			name: "production environment",
			cfg: &config.Config{
				App: config.AppConfig{
					Environment: "production",
				},
				Log: config.LogConfig{
					Level: "info",
					Debug: false,
				},
			},
			wantErr: false,
		},
		{
			name: "empty environment defaults to development",
			cfg: &config.Config{
				App: config.AppConfig{},
				Log: config.LogConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "empty log level defaults to info",
			cfg: &config.Config{
				App: config.AppConfig{
					Environment: "development",
				},
				Log: config.LogConfig{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.InitializeLogger(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, log)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, log)
			}
		})
	}
}

func TestCustomLogger_Methods(t *testing.T) {
	// Create a temporary log file
	tmpLogFile := "test.log"
	defer os.Remove(tmpLogFile)

	// Create a development logger
	log, err := logger.NewDevelopmentLogger("debug")
	require.NoError(t, err)
	require.NotNil(t, log)

	// Test all logging methods
	t.Run("Info", func(_ *testing.T) {
		log.Info("test info message", "key", "value")
	})

	t.Run("Error", func(_ *testing.T) {
		log.Error("test error message", "key", "value")
	})

	t.Run("Debug", func(_ *testing.T) {
		log.Debug("test debug message", "key", "value")
	})

	t.Run("Warn", func(_ *testing.T) {
		log.Warn("test warn message", "key", "value")
	})

	t.Run("Fatal with recovery", func(_ *testing.T) {
		defer func() {
			_ = recover() // Expected panic from Fatal
		}()
		log.Fatal("test fatal message", "key", "value")
	})

	t.Run("Printf", func(_ *testing.T) {
		log.Printf("test printf %s", "value")
	})

	t.Run("Errorf", func(_ *testing.T) {
		log.Errorf("test errorf %s", "value")
	})

	t.Run("Sync", func(t *testing.T) {
		syncErr := log.Sync()
		// Ignore sync errors as they're expected when writing to console
		if syncErr != nil {
			t.Log("Sync() error (expected):", syncErr)
		}
	})
}

func TestLoggerContext(t *testing.T) {
	// Create a context with logger
	zapLogger := zap.NewExample()
	ctx := logger.WithContext(t.Context(), zapLogger)

	// Test retrieving logger from context
	t.Run("FromContext with logger", func(t *testing.T) {
		retrieved := logger.FromContext(ctx)
		assert.NotNil(t, retrieved)
		assert.Equal(t, zapLogger, retrieved)
	})

	t.Run("FromContext without logger", func(t *testing.T) {
		emptyCtx := t.Context()
		retrieved := logger.FromContext(emptyCtx)
		assert.NotNil(t, retrieved) // Should return no-op logger
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		expected  zapcore.Level
		expectErr bool
	}{
		{
			name:      "debug level",
			levelStr:  "debug",
			expected:  zapcore.DebugLevel,
			expectErr: false,
		},
		{
			name:      "info level",
			levelStr:  "info",
			expected:  zapcore.InfoLevel,
			expectErr: false,
		},
		{
			name:      "warn level",
			levelStr:  "warn",
			expected:  zapcore.WarnLevel,
			expectErr: false,
		},
		{
			name:      "error level",
			levelStr:  "error",
			expected:  zapcore.ErrorLevel,
			expectErr: false,
		},
		{
			name:      "invalid level",
			levelStr:  "invalid",
			expected:  zapcore.DebugLevel,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := logger.ParseLogLevel(tt.levelStr)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestConvertToZapFields(t *testing.T) {
	tests := []struct {
		name          string
		fields        []interface{}
		expectedCount int
	}{
		{
			name:          "empty fields",
			fields:        []interface{}{},
			expectedCount: 0,
		},
		{
			name:          "single key-value pair",
			fields:        []interface{}{"key", "value"},
			expectedCount: 1,
		},
		{
			name:          "multiple key-value pairs",
			fields:        []interface{}{"key1", "value1", "key2", "value2"},
			expectedCount: 2,
		},
		{
			name:          "odd number of fields",
			fields:        []interface{}{"key1", "value1", "extra"},
			expectedCount: 2,
		},
		{
			name:          "non-string key",
			fields:        []interface{}{123, "value"},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			fields := logger.ConvertToZapFields(tt.fields)
			assert.Len(t, fields, tt.expectedCount)
		})
	}
}

func TestCustomLogger_LoggingWithFields(t *testing.T) {
	log, err := logger.NewDevelopmentLogger("debug")
	require.NoError(t, err)

	tests := []struct {
		name   string
		fields []interface{}
	}{
		{
			name:   "no fields",
			fields: []interface{}{},
		},
		{
			name:   "single key-value pair",
			fields: []interface{}{"key", "value"},
		},
		{
			name:   "multiple key-value pairs",
			fields: []interface{}{"key1", "value1", "key2", "value2"},
		},
		{
			name:   "odd number of fields",
			fields: []interface{}{"key1", "value1", "extra"},
		},
		{
			name:   "non-string key",
			fields: []interface{}{123, "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Test all logging methods with fields
			log.Info("test message", tt.fields...)
			log.Error("test message", tt.fields...)
			log.Debug("test message", tt.fields...)
			log.Warn("test message", tt.fields...)
		})
	}
}
