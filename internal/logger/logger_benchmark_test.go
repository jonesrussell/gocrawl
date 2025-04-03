package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
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

	logConfig := &logger.Config{
		Level:       logger.InfoLevel,
		Development: false,
	}
	log := createLogger(zapLogger, logConfig)

	b.Run("Info", func(b *testing.B) {
		for _ = range b.N {
			log.Info("benchmark message", "key", "value")
		}
	})

	b.Run("InfoWithFields", func(b *testing.B) {
		for _ = range b.N {
			log.Info("benchmark message",
				"key1", "value1",
				"key2", "value2",
				"key3", "value3",
			)
		}
	})

	b.Run("With", func(b *testing.B) {
		for _ = range b.N {
			child := log.With("key", "value")
			child.Info("benchmark message")
		}
	})

	b.Run("Error", func(b *testing.B) {
		for _ = range b.N {
			log.Error("benchmark error", "error", "test error")
		}
	})
}
