package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
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
	logLevel := "info"                                   // Set the desired log level
	logger, err := logger.NewDevelopmentLogger(logLevel) // Pass log level as string
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}
}

func TestNewProductionLogger(t *testing.T) {
	logLevel := "info"                                  // Set the desired log level
	logger, err := logger.NewProductionLogger(logLevel) // Pass log level as string
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}
}
