// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface defines the logger interface.
type Interface interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	With(fields ...any) Interface
}

// Logger implements the Interface.
type Logger struct {
	zapLogger *zap.Logger
}

var (
	// defaultLogger is the singleton logger instance
	defaultLogger *Logger
)

// New creates a new logger instance.
func New(config *Config) (Interface, error) {
	if defaultLogger != nil {
		return defaultLogger, nil
	}

	// Set default values
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Encoding == "" {
		config.Encoding = "console"
	}
	if len(config.OutputPaths) == 0 {
		config.OutputPaths = []string{"stdout"}
	}

	// Create development encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set output
	output := zapcore.AddSync(os.Stdout)

	// Create level enabler
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder based on encoding setting
	var encoder zapcore.Encoder
	if config.Encoding == "console" {
		if config.EnableColor {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		output,
		level,
	)

	// Create logger with options
	opts := []zap.Option{
		zap.AddCaller(),
	}

	// Add development mode options
	if config.Development {
		opts = append(opts,
			zap.Development(),
			zap.AddStacktrace(zapcore.WarnLevel),
		)
	} else {
		opts = append(opts,
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	}

	// Create logger
	zapLogger := zap.New(core, opts...)
	defaultLogger = &Logger{zapLogger: zapLogger}

	// Log configuration details
	defaultLogger.Debug("Logger configuration",
		"level", config.Level,
		"development", config.Development,
		"encoding", config.Encoding,
		"output_paths", config.OutputPaths,
		"enable_color", config.EnableColor,
		"zap_level", level.String(),
		"zap_options", fmt.Sprintf("%+v", opts))

	return defaultLogger, nil
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a new logger with the given fields.
func (l *Logger) With(fields ...any) Interface {
	return &Logger{
		zapLogger: l.zapLogger.With(toZapFields(fields)...),
	}
}

// fieldPairSize represents the number of elements in a key-value pair.
const fieldPairSize = 2

// toZapFields converts a list of any fields to zap.Field.
func toZapFields(fields []any) []zap.Field {
	if len(fields)%fieldPairSize != 0 {
		return []zap.Field{zap.Error(ErrInvalidFields)}
	}

	zapFields := make([]zap.Field, 0, len(fields)/fieldPairSize)
	for i := 0; i < len(fields); i += fieldPairSize {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(ErrInvalidFields)}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}

// NewFromConfig creates a new logger instance from Viper configuration
func NewFromConfig(v *viper.Viper) (Interface, error) {
	// Get logger configuration from Viper
	logConfig := &Config{
		Level:       Level(v.GetString("logger.level")),
		Development: v.GetBool("logger.development"),
		Encoding:    v.GetString("logger.encoding"),
		OutputPaths: v.GetStringSlice("logger.output_paths"),
		EnableColor: v.GetBool("logger.enable_color"),
	}

	// If no output paths specified, default to stdout
	if len(logConfig.OutputPaths) == 0 {
		logConfig.OutputPaths = DefaultOutputPaths
	}

	// Create logger with configuration
	log, err := New(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Log the current log level
	log.Debug("Logger initialized",
		"level", logConfig.Level,
		"development", logConfig.Development,
		"encoding", logConfig.Encoding,
		"output_paths", logConfig.OutputPaths,
		"enable_color", logConfig.EnableColor)

	return log, nil
}

// GetZapLogger returns the underlying zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// NewWithZap creates a new logger instance and returns both Interface and *zap.Logger
func NewWithZap(config *Config) (Interface, *zap.Logger, error) {
	logger, err := New(config)
	if err != nil {
		return nil, nil, err
	}

	if l, ok := logger.(*Logger); ok {
		return logger, l.zapLogger, nil
	}

	return logger, nil, fmt.Errorf("failed to get underlying zap logger")
}
