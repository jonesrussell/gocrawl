package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		params  logger.Params
		wantErr bool
	}{
		{
			name: "development logger with debug",
			params: logger.Params{
				Debug:  true,
				Level:  zapcore.DebugLevel.String(),
				AppEnv: "development",
			},
			wantErr: false,
		},
		{
			name: "production logger",
			params: logger.Params{
				Debug:  false,
				Level:  zapcore.InfoLevel.String(),
				AppEnv: "production",
			},
			wantErr: false,
		},
		{
			name: "default environment",
			params: logger.Params{
				Debug: true,
				Level: zapcore.DebugLevel.String(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a zap logger with the test parameters
			config := zap.NewDevelopmentConfig()
			level, err := zapcore.ParseLevel(tt.params.Level)
			if err != nil {
				t.Errorf("Failed to parse log level: %v", err)
				return
			}
			config.Level = zap.NewAtomicLevelAt(level)
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our logger
			log := &logger.ZapLogger{Logger: zapLogger}

			// Test logging methods
			log.Info("test info message", "key", "value")
			log.Error("test error message", "key", "value")
			log.Debug("test debug message", "key", "value")
			log.Warn("test warn message", "key", "value")
			log.Errorf("test error format %s", "value")

			// Test Sync
			if syncErr := log.Sync(); syncErr != nil {
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
			level, err := zapcore.ParseLevel(tt.levelStr)
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

			// Create our logger
			log := &logger.ZapLogger{Logger: zapLogger}

			// Test logging methods
			log.Info("test info message", "key", "value")
			log.Error("test error message", "key", "value")
			log.Debug("test debug message", "key", "value")
			log.Warn("test warn message", "key", "value")
			log.Errorf("test error format %s", "value")
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
			level, err := zapcore.ParseLevel(tt.levelStr)
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

			// Create our logger
			log := &logger.ZapLogger{Logger: zapLogger}

			// Test logging methods
			log.Info("test info message", "key", "value")
			log.Error("test error message", "key", "value")
			log.Warn("test warn message", "key", "value")
			log.Errorf("test error format %s", "value")
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	log := logger.NewTestLogger()
	defer log.Sync()

	// Test all logging methods
	log.Debug("debug message", "key", "value")
	log.Info("info message", "key", "value")
	log.Warn("warn message", "key", "value")
	log.Error("error message", "key", "value")
	log.Printf("printf message %s", "value")
	log.Errorf("errorf message %s", "value")
}

func TestNoOpLogger(t *testing.T) {
	log := logger.NewNoOp()

	// These should not panic
	log.Debug("debug message", "key", "value")
	log.Info("info message", "key", "value")
	log.Warn("warn message", "key", "value")
	log.Error("error message", "key", "value")
	log.Printf("printf message %s", "value")
	log.Errorf("errorf message %s", "value")
	log.Sync()
}
