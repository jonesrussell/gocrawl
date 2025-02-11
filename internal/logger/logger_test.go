package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewCustomLogger(t *testing.T) {
	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name: "development logger with debug",
			params: Params{
				Debug:  true,
				Level:  zapcore.DebugLevel,
				AppEnv: "development",
			},
			wantErr: false,
		},
		{
			name: "production logger",
			params: Params{
				Debug:  false,
				Level:  zapcore.InfoLevel,
				AppEnv: "production",
			},
			wantErr: false,
		},
		{
			name: "default environment",
			params: Params{
				Debug: true,
				Level: zapcore.DebugLevel,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewDevelopmentLogger(tt.params)
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
			if err := logger.Sync(); err != nil {
				// Ignore sync errors as they're expected when writing to console
				t.Log("Sync() error (expected):", err)
			}
		})
	}
}

func TestConvertToZapFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []interface{}
		want   int // number of expected zap fields
	}{
		{
			name:   "empty fields",
			fields: []interface{}{},
			want:   0,
		},
		{
			name:   "single string context",
			fields: []interface{}{"context message"},
			want:   1,
		},
		{
			name:   "key-value pairs",
			fields: []interface{}{"key1", "value1", "key2", 42},
			want:   2,
		},
		{
			name:   "odd number of fields",
			fields: []interface{}{"key1", "value1", "extra"},
			want:   2, // key-value pair plus context
		},
		{
			name:   "non-string keys",
			fields: []interface{}{42, "value1"},
			want:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToZapFields(tt.fields)
			if len(got) != tt.want {
				t.Errorf("convertToZapFields() got %v fields, want %v", len(got), tt.want)
			}
		})
	}
}
