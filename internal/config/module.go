// Package config provides configuration management for the GoCrawl application.
// This file specifically handles dependency injection and module initialization
// using the fx framework.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// setupConfig holds configuration setup dependencies
type setupConfig struct {
	fx.In

	Logger Logger `optional:"true"`
}

// defaultLogger provides a no-op logger when no logger is injected
type defaultLogger struct{}

func (l defaultLogger) Info(msg string, fields ...Field) {}
func (l defaultLogger) Warn(msg string, fields ...Field) {}

// New creates a new Config instance
func New(log Logger) (Interface, error) {
	// Initialize Viper if no config file is set
	if viper.GetViper().ConfigFileUsed() == "" {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	// Check for test environment
	if os.Getenv("GOCRAWL_APP_ENVIRONMENT") == "test" {
		// Set test-specific defaults
		viper.SetDefault("app.environment", "test")
		viper.SetDefault("app.name", "gocrawl-test")
		viper.SetDefault("app.version", "0.0.1")
		viper.SetDefault("app.debug", false)
		viper.SetDefault("log.level", "debug")
		viper.SetDefault("log.debug", true)
		viper.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
		viper.SetDefault("elasticsearch.index_name", "test-index")
		viper.SetDefault("elasticsearch.api_key", "test_api_key")
		viper.SetDefault("elasticsearch.retry.enabled", true)
		viper.SetDefault("elasticsearch.retry.initial_wait", "1s")
		viper.SetDefault("elasticsearch.retry.max_wait", "5s")
		viper.SetDefault("elasticsearch.retry.max_retries", 3)
		viper.SetDefault("server.address", ":8080")
		viper.SetDefault("server.security.enabled", true)
		viper.SetDefault("server.security.api_key", "test_api_key")
		viper.SetDefault("crawler.base_url", "http://test.example.com")
		viper.SetDefault("crawler.max_depth", 2)
		viper.SetDefault("crawler.rate_limit", "2s")
		viper.SetDefault("crawler.parallelism", 2)
	}

	cfg := &Config{
		logger: log,
	}
	if err := cfg.load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// loadEnvFile loads an environment file if it exists and logs any errors
func loadEnvFile(log Logger, envFile string) error {
	if err := godotenv.Load(envFile); err != nil {
		if !os.IsNotExist(err) {
			log.Warn("Error loading environment file",
				String("file", envFile),
				Error(err))
			return err
		}
		return nil
	}
	log.Info("Loaded environment file",
		String("file", envFile))
	return nil
}

// SetupConfig initializes the configuration system with an optional environment file
func SetupConfig(log Logger, envFile string) error {
	// Initialize Viper configuration
	if err := setupViper(log); err != nil {
		return fmt.Errorf("failed to setup Viper: %w", err)
	}

	// Load environment file
	if envFile != "" {
		if err := loadEnvFile(log, envFile); err != nil {
			return fmt.Errorf("failed to load environment file: %w", err)
		}
	} else {
		// Load default .env file if it exists
		if err := loadEnvFile(log, ".env"); err != nil {
			return fmt.Errorf("failed to load default .env file: %w", err)
		}
	}

	// Bind environment variables
	if err := bindEnvs(defaultEnvBindings()); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	return nil
}

// provideConfig creates and initializes the configuration provider
func provideConfig(envFile string) func(setupConfig) (Interface, error) {
	return func(setup setupConfig) (Interface, error) {
		// Use injected logger or fallback to default
		log := setup.Logger
		if log == nil {
			log = defaultLogger{}
		}

		if err := SetupConfig(log, envFile); err != nil {
			return nil, err
		}
		return New(log)
	}
}

// Module provides the config module and its dependencies using fx.
// It sets up the configuration providers that can be used throughout
// the application for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		fx.Annotate(
			provideConfig(""),
			fx.As(new(Interface)),
			fx.ResultTags(`optional:"true"`),
		),
	),
)
