package logger

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkLogger(b *testing.B) {
	// Create a test logger
	config := zap.NewProductionConfig()
	zapLogger, err := config.Build()
	if err != nil {
		b.Fatalf("Failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()

	logConfig := &Config{
		Level:       DebugLevel,
		Development: true,
	}
	log := &logger{
		zapLogger: zapLogger,
		config:    logConfig,
	}

	// Run benchmarks
	b.Run("Debug", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Debug("test message", "key", "value")
		}
	})

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Info("test message", "key", "value")
		}
	})

	b.Run("Warn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Warn("test message", "key", "value")
		}
	})

	b.Run("Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Error("test message", "key", "value")
		}
	})

	b.Run("With", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.With("key", "value")
		}
	})

	b.Run("NewFxLogger", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.NewFxLogger()
		}
	})
}
