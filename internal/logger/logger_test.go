package logger_test

import (
	"fmt"
	"testing"

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
			// Create a zap logger with the test parameters
			config := zap.NewDevelopmentConfig()
			config.Level = zap.NewAtomicLevelAt(tt.params.Level)
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our custom logger
			customLogger, err := logger.NewCustomLogger(zapLogger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustomLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if customLogger == nil {
				t.Error("NewCustomLogger() returned nil customLogger")
				return
			}

			// Test logging methods
			customLogger.Info("test info message", "key", "value")
			customLogger.Error("test error message", "key", "value")
			customLogger.Debug("test debug message", "key", "value")
			customLogger.Warn("test warn message", "key", "value")
			customLogger.Errorf("test error format %s", "value")

			// Test GetZapLogger
			if customLogger.GetZapLogger() == nil {
				t.Error("GetZapLogger() returned nil")
			}

			// Test Sync
			if syncErr := customLogger.Sync(); syncErr != nil {
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
			name:      "invalid level",
			levelStr:  "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a zap logger with development config
			config := zap.NewDevelopmentConfig()
			level, err := logger.ParseLogLevel(tt.levelStr)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseLogLevel() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if tt.expectErr {
				return
			}
			config.Level = zap.NewAtomicLevelAt(level)
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our custom logger
			customLogger, err := logger.NewCustomLogger(zapLogger)
			if err != nil {
				t.Errorf("NewCustomLogger() error = %v", err)
				return
			}
			if customLogger == nil {
				t.Error("NewCustomLogger() returned nil")
				return
			}

			// Test logging methods
			customLogger.Info("test info message", "key", "value")
			customLogger.Error("test error message", "key", "value")
			customLogger.Debug("test debug message", "key", "value")
			customLogger.Warn("test warn message", "key", "value")
			customLogger.Errorf("test error format %s", "value")
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
			// Create a zap logger with production config
			config := zap.NewProductionConfig()
			level, err := logger.ParseLogLevel(tt.levelStr)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseLogLevel() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if tt.expectErr {
				return
			}
			config.Level = zap.NewAtomicLevelAt(level)
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our custom logger
			customLogger, err := logger.NewCustomLogger(zapLogger)
			if err != nil {
				t.Errorf("NewCustomLogger() error = %v", err)
				return
			}
			if customLogger == nil {
				t.Error("NewCustomLogger() returned nil")
				return
			}

			// Test logging methods
			customLogger.Info("test info message", "key", "value")
			customLogger.Error("test error message", "key", "value")
			customLogger.Warn("test warn message", "key", "value")
			customLogger.Errorf("test error format %s", "value")
		})
	}
}

// createTestLogger creates a logger with the given configuration for testing
func createTestLogger(t *testing.T, levelStr string, appEnv string) (*logger.CustomLogger, error) {
	var config zap.Config
	if appEnv == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	level, err := logger.ParseLogLevel(levelStr)
	if err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(level)
	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return logger.NewCustomLogger(zapLogger)
}

func TestInitializeLogger(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		appEnv    string
		expectErr bool
	}{
		{
			name:      "development environment",
			levelStr:  "debug",
			appEnv:    "development",
			expectErr: false,
		},
		{
			name:      "production environment",
			levelStr:  "info",
			appEnv:    "production",
			expectErr: false,
		},
		{
			name:      "invalid level",
			levelStr:  "invalid",
			appEnv:    "development",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customLogger, err := createTestLogger(t, tt.levelStr, tt.appEnv)
			if (err != nil) != tt.expectErr {
				t.Errorf("createTestLogger() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if tt.expectErr {
				return
			}
			if customLogger == nil {
				t.Error("createTestLogger() returned nil")
				return
			}

			// Test logging methods
			customLogger.Info("test info message", "key", "value")
			customLogger.Error("test error message", "key", "value")
			if tt.appEnv == "development" {
				customLogger.Debug("test debug message", "key", "value")
			}
			customLogger.Warn("test warn message", "key", "value")
			customLogger.Errorf("test error format %s", "value")
		})
	}
}

func TestCustomLogger_Methods(t *testing.T) {
	// Create a zap logger with development config
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zapLogger, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()

	// Create our custom logger
	customLogger, err := logger.NewCustomLogger(zapLogger)
	if err != nil {
		t.Fatalf("Failed to create custom logger: %v", err)
	}

	// Test all logging methods
	t.Run("Info", func(t *testing.T) {
		customLogger.Info("test info message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		customLogger.Error("test error message", "key", "value")
	})

	t.Run("Debug", func(t *testing.T) {
		customLogger.Debug("test debug message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		customLogger.Warn("test warn message", "key", "value")
	})

	t.Run("Printf", func(t *testing.T) {
		customLogger.Printf("test printf message %s", "value")
	})

	t.Run("Errorf", func(t *testing.T) {
		customLogger.Errorf("test errorf message %s", "value")
	})

	t.Run("Fatal", func(t *testing.T) {
		// Set up a fatal hook to prevent actual fatal
		customLogger.SetFatalHook(func(entry zapcore.Entry) error {
			return nil
		})
		customLogger.Fatal("test fatal message", "key", "value")
	})

	t.Run("Sync", func(t *testing.T) {
		syncErr := customLogger.Sync()
		if syncErr != nil {
			// Ignore sync errors as they're expected when writing to console
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
		fields        []any
		expectedCount int
	}{
		{
			name:          "empty fields",
			fields:        []any{},
			expectedCount: 0,
		},
		{
			name:          "single key-value pair",
			fields:        []any{"key", "value"},
			expectedCount: 1,
		},
		{
			name:          "multiple key-value pairs",
			fields:        []any{"key1", "value1", "key2", "value2"},
			expectedCount: 2,
		},
		{
			name:          "odd number of fields",
			fields:        []any{"key1", "value1", "extra"},
			expectedCount: 2,
		},
		{
			name:          "non-string key",
			fields:        []any{123, "value"},
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
	// Create a zap logger with development config
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zapLogger, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()

	// Create our custom logger
	customLogger, err := logger.NewCustomLogger(zapLogger)
	if err != nil {
		t.Fatalf("Failed to create custom logger: %v", err)
	}

	// Test logging with various field types
	tests := []struct {
		name   string
		msg    string
		fields []any
	}{
		{
			name:   "string fields",
			msg:    "test message",
			fields: []any{"key1", "value1", "key2", "value2"},
		},
		{
			name:   "numeric fields",
			msg:    "test message",
			fields: []any{"int", 42, "float", 3.14},
		},
		{
			name:   "boolean fields",
			msg:    "test message",
			fields: []any{"bool", true, "bool2", false},
		},
		{
			name:   "nil fields",
			msg:    "test message",
			fields: []any{"nil", nil},
		},
		{
			name:   "mixed fields",
			msg:    "test message",
			fields: []any{"string", "value", "int", 42, "bool", true, "nil", nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customLogger.Info(tt.msg, tt.fields...)
			customLogger.Error(tt.msg, tt.fields...)
			customLogger.Debug(tt.msg, tt.fields...)
			customLogger.Warn(tt.msg, tt.fields...)
		})
	}
}
