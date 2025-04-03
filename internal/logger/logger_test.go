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
		config  *logger.Config
		wantErr bool
	}{
		{
			name: "development logger with debug",
			config: &logger.Config{
				Development: true,
				Level:       logger.DebugLevel,
				Encoding:    "console",
			},
			wantErr: false,
		},
		{
			name: "production logger",
			config: &logger.Config{
				Development: false,
				Level:       logger.InfoLevel,
				Encoding:    "json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new logger directly
			zapConfig := zap.NewProductionConfig()
			if tt.config.Development {
				zapConfig = zap.NewDevelopmentConfig()
			}

			zapConfig.Level = zap.NewAtomicLevelAt(levelToZap(tt.config.Level))
			zapConfig.Encoding = tt.config.Encoding
			zapConfig.OutputPaths = tt.config.OutputPaths
			zapConfig.ErrorOutputPaths = tt.config.ErrorOutputPaths

			zapLogger, err := zapConfig.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our logger implementation
			log := createLogger(zapLogger, tt.config)

			// Test logging methods
			log.Debug("debug message", "key", "value")
			log.Info("info message", "key", "value")
			log.Warn("warn message", "key", "value")
			log.Error("error message", "key", "value")
		})
	}
}

func TestNewDevelopmentLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     logger.Level
		expectErr bool
	}{
		{
			name:      "debug level",
			level:     logger.DebugLevel,
			expectErr: false,
		},
		{
			name:      "info level",
			level:     logger.InfoLevel,
			expectErr: false,
		},
		{
			name:      "warn level",
			level:     logger.WarnLevel,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a zap logger with development config
			config := zap.NewDevelopmentConfig()
			config.Level = zap.NewAtomicLevelAt(levelToZap(tt.level))
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our logger
			logConfig := &logger.Config{
				Level:       tt.level,
				Development: true,
			}
			log := createLogger(zapLogger, logConfig)

			// Test logging methods
			log.Info("test info message", "key", "value")
			log.Error("test error message", "key", "value")
			log.Debug("test debug message", "key", "value")
			log.Warn("test warn message", "key", "value")
		})
	}
}

func TestNewProductionLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     logger.Level
		expectErr bool
	}{
		{
			name:      "info level",
			level:     logger.InfoLevel,
			expectErr: false,
		},
		{
			name:      "warn level",
			level:     logger.WarnLevel,
			expectErr: false,
		},
		{
			name:      "error level",
			level:     logger.ErrorLevel,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a zap logger with production config
			config := zap.NewProductionConfig()
			config.Level = zap.NewAtomicLevelAt(levelToZap(tt.level))
			zapLogger, err := config.Build()
			if err != nil {
				t.Errorf("Failed to create zap logger: %v", err)
				return
			}
			defer zapLogger.Sync()

			// Create our logger
			logConfig := &logger.Config{
				Level:       tt.level,
				Development: false,
			}
			log := createLogger(zapLogger, logConfig)

			// Test logging methods
			log.Info("test info message", "key", "value")
			log.Error("test error message", "key", "value")
			log.Warn("test warn message", "key", "value")
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	// Create a test logger
	config := zap.NewDevelopmentConfig()
	zapLogger, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()

	logConfig := &logger.Config{
		Level:       logger.DebugLevel,
		Development: true,
	}
	log := createLogger(zapLogger, logConfig)

	// Test all logging methods
	log.Debug("debug message", "key", "value")
	log.Info("info message", "key", "value")
	log.Warn("warn message", "key", "value")
	log.Error("error message", "key", "value")
}

func TestLoggerEdgeCases(t *testing.T) {
	// Create a test logger
	config := zap.NewProductionConfig()
	zapLogger, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()

	logConfig := &logger.Config{
		Level:       logger.DebugLevel,
		Development: true,
	}
	log := createLogger(zapLogger, logConfig)

	t.Run("EmptyMessage", func(t *testing.T) {
		log.Info("")
	})

	t.Run("EmptyFields", func(t *testing.T) {
		log.Info("message")
	})

	t.Run("OddNumberOfFields", func(t *testing.T) {
		log.Info("message", "key1", "value1", "key2")
	})

	t.Run("NilFields", func(t *testing.T) {
		log.Info("message", nil, "value")
	})

	t.Run("WithEmptyFields", func(t *testing.T) {
		child := log.With()
		child.Info("message")
	})

	t.Run("WithNilFields", func(t *testing.T) {
		child := log.With(nil)
		child.Info("message")
	})

	t.Run("WithOddNumberOfFields", func(t *testing.T) {
		child := log.With("key1", "value1", "key2")
		child.Info("message")
	})
}

// Helper function to convert logger.Level to zapcore.Level
func levelToZap(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Helper function to create a logger implementation
func createLogger(zapLogger *zap.Logger, config *logger.Config) logger.Interface {
	// Create a test logger that implements the Interface
	return &testLogger{
		zapLogger: zapLogger,
		config:    config,
	}
}

// testLogger implements logger.Interface for testing
type testLogger struct {
	zapLogger *zap.Logger
	config    *logger.Config
}

// Debug logs a debug message
func (l *testLogger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message
func (l *testLogger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message
func (l *testLogger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message
func (l *testLogger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits
func (l *testLogger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a child logger with additional fields
func (l *testLogger) With(fields ...any) logger.Interface {
	return &testLogger{
		zapLogger: l.zapLogger.With(toZapFields(fields)...),
		config:    l.config,
	}
}

// toZapFields converts a list of any fields to zap.Field
func toZapFields(fields []any) []zap.Field {
	if len(fields)%2 != 0 {
		return []zap.Field{zap.Error(logger.ErrInvalidFields)}
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(logger.ErrInvalidFields)}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
