// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/cmd/common"
	crawlcmd "github.com/jonesrussell/gocrawl/cmd/crawl"
	httpdcmd "github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/index"
	"github.com/jonesrussell/gocrawl/cmd/scheduler"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
)

var (
	// cfgFile holds the path to the configuration file.
	// It can be set via the --config flag or defaults to config.yaml.
	cfgFile string

	// Debug enables debug mode for all commands
	Debug bool

	// rootCmd represents the root command for the GoCrawl CLI.
	// It serves as the base command that all subcommands are attached to.
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler for collecting and processing content",
		Long: `gocrawl is a web crawler that helps you collect and process content from various sources.
It provides a flexible and extensible framework for building custom crawlers.`,
	}
)

const (
	defaultCrawlerMaxDepth      = 10
	defaultCrawlerMaxRetries    = 3
	defaultStorageBatchSize     = 100
	defaultElasticsearchRetries = 3
	logLevelDebug               = "debug"
	logLevelInfo                = "info"
	logLevelWarn                = "warn"
	logLevelError               = "error"
	// DefaultBulkSize is the default number of documents to bulk index
	DefaultBulkSize = 1000
)

// bindAppEnvVars binds app-related environment variables
func bindAppEnvVars() error {
	if err := viper.BindEnv("app.name", "APP_NAME"); err != nil {
		return fmt.Errorf("failed to bind app.name: %w", err)
	}
	if err := viper.BindEnv("app.environment", "APP_ENV"); err != nil {
		return fmt.Errorf("failed to bind app.environment: %w", err)
	}
	if err := viper.BindEnv("app.debug", "APP_DEBUG"); err != nil {
		return fmt.Errorf("failed to bind app.debug: %w", err)
	}
	return nil
}

// bindServerEnvVars binds server-related environment variables
func bindServerEnvVars() error {
	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return fmt.Errorf("failed to bind server.port: %w", err)
	}
	if err := viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind server.read_timeout: %w", err)
	}
	if err := viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind server.write_timeout: %w", err)
	}
	if err := viper.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind server.idle_timeout: %w", err)
	}
	return nil
}

// bindElasticsearchEnvVars binds Elasticsearch-related environment variables
func bindElasticsearchEnvVars() error {
	if err := viper.BindEnv("elasticsearch.hosts", "ELASTICSEARCH_HOSTS"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.hosts: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.api_key", "ELASTICSEARCH_API_KEY"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.api_key: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.index_prefix", "ELASTICSEARCH_INDEX_PREFIX"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.index_prefix: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.max_retries", "ELASTICSEARCH_MAX_RETRIES"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.max_retries: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.retry_initial_wait", "ELASTICSEARCH_RETRY_INITIAL_WAIT"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.retry_initial_wait: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.retry_max_wait", "ELASTICSEARCH_RETRY_MAX_WAIT"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.retry_max_wait: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.discover_nodes", "ELASTICSEARCH_DISCOVER_NODES"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.discover_nodes: %w", err)
	}
	if err := viper.BindEnv(
		"elasticsearch.tls_insecure_skip_verify",
		"ELASTICSEARCH_TLS_INSECURE_SKIP_VERIFY",
	); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.tls_insecure_skip_verify: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.ca_fingerprint", "ELASTICSEARCH_CA_FINGERPRINT"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.ca_fingerprint: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.username", "ELASTICSEARCH_USERNAME"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.username: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.password", "ELASTICSEARCH_PASSWORD"); err != nil {
		return fmt.Errorf("failed to bind elasticsearch.password: %w", err)
	}
	return nil
}

// bindCrawlerEnvVars binds crawler-related environment variables
func bindCrawlerEnvVars() error {
	if err := viper.BindEnv("crawler.max_depth", "CRAWLER_MAX_DEPTH"); err != nil {
		return fmt.Errorf("failed to bind crawler.max_depth: %w", err)
	}
	if err := viper.BindEnv("crawler.parallelism", "CRAWLER_PARALLELISM"); err != nil {
		return fmt.Errorf("failed to bind crawler.parallelism: %w", err)
	}
	if err := viper.BindEnv("crawler.max_age", "CRAWLER_MAX_AGE"); err != nil {
		return fmt.Errorf("failed to bind crawler.max_age: %w", err)
	}
	if err := viper.BindEnv("crawler.rate_limit", "CRAWLER_RATE_LIMIT"); err != nil {
		return fmt.Errorf("failed to bind crawler.rate_limit: %w", err)
	}
	if err := viper.BindEnv("crawler.debugger.enabled", "CRAWLER_DEBUGGER_ENABLED"); err != nil {
		return fmt.Errorf("failed to bind crawler.debugger.enabled: %w", err)
	}
	if err := viper.BindEnv("crawler.debugger.level", "CRAWLER_DEBUGGER_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind crawler.debugger.level: %w", err)
	}
	if err := viper.BindEnv("crawler.debugger.format", "CRAWLER_DEBUGGER_FORMAT"); err != nil {
		return fmt.Errorf("failed to bind crawler.debugger.format: %w", err)
	}
	if err := viper.BindEnv("crawler.debugger.output", "CRAWLER_DEBUGGER_OUTPUT"); err != nil {
		return fmt.Errorf("failed to bind crawler.debugger.output: %w", err)
	}
	if err := viper.BindEnv("crawler.debugger.file", "CRAWLER_DEBUGGER_FILE"); err != nil {
		return fmt.Errorf("failed to bind crawler.debugger.file: %w", err)
	}
	return nil
}

// bindLoggerEnvVars binds logger-related environment variables
func bindLoggerEnvVars() error {
	if err := viper.BindEnv("logger.level", "LOG_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind logger.level: %w", err)
	}
	if err := viper.BindEnv("logger.format", "LOG_FORMAT"); err != nil {
		return fmt.Errorf("failed to bind logger.format: %w", err)
	}
	return nil
}

// bindServerSecurityEnvVars binds server security-related environment variables
func bindServerSecurityEnvVars() error {
	if err := viper.BindEnv("server.port", "GOCRAWL_PORT"); err != nil {
		return fmt.Errorf("failed to bind server.port: %w", err)
	}
	if err := viper.BindEnv("server.api_key", "GOCRAWL_API_KEY"); err != nil {
		return fmt.Errorf("failed to bind server.api_key: %w", err)
	}
	return nil
}

// bindEnvVars binds all environment variables to configuration keys
func bindEnvVars() error {
	if err := bindAppEnvVars(); err != nil {
		return err
	}
	if err := bindServerEnvVars(); err != nil {
		return err
	}
	if err := bindElasticsearchEnvVars(); err != nil {
		return err
	}
	if err := bindCrawlerEnvVars(); err != nil {
		return err
	}
	if err := bindLoggerEnvVars(); err != nil {
		return err
	}
	if err := bindServerSecurityEnvVars(); err != nil {
		return err
	}
	return nil
}

// setDefaults sets default values for configuration keys
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "gocrawl")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)

	// Logger defaults
	viper.SetDefault("logger.level", "debug")
	viper.SetDefault("logger.format", "console")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.enable_color", true)

	// Crawler defaults
	viper.SetDefault("crawler.max_depth", defaultCrawlerMaxDepth)
	viper.SetDefault("crawler.max_retries", defaultCrawlerMaxRetries)
	viper.SetDefault("crawler.rate_limit", "1s")
	viper.SetDefault("crawler.timeout", "30s")
	viper.SetDefault("crawler.user_agent", "GoCrawl/1.0")

	// Storage defaults
	viper.SetDefault("storage.type", "elasticsearch")
	viper.SetDefault("storage.batch_size", defaultStorageBatchSize)
	viper.SetDefault("storage.flush_interval", "5s")

	// Elasticsearch defaults
	viper.SetDefault("elasticsearch.addresses", []string{"https://localhost:9200"})
	viper.SetDefault("elasticsearch.index_name", "gocrawl")
	viper.SetDefault("elasticsearch.retry.enabled", true)
	viper.SetDefault("elasticsearch.retry.initial_wait", "1s")
	viper.SetDefault("elasticsearch.retry.max_wait", "5s")
	viper.SetDefault("elasticsearch.retry.max_retries", defaultElasticsearchRetries)
	viper.SetDefault("elasticsearch.bulk_size", DefaultBulkSize)
	viper.SetDefault("elasticsearch.flush_interval", "1s")
	viper.SetDefault("elasticsearch.tls.insecure_skip_verify", true)
}

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() error {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		// .env file not found is not an error - we'll use environment variables
	}
	return nil
}

// loadConfigFile loads the configuration file from the specified path or default locations
func loadConfigFile() error {
	if cfgFile != "" {
		absPath, err := filepath.Abs(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for config file: %w", err)
		}
		cfgFile = absPath
		viper.SetConfigFile(cfgFile)
	} else {
		// Set up default configuration paths and name
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		// Search paths in order of priority
		viper.AddConfigPath(".")                             // Current directory
		viper.AddConfigPath(filepath.Join(home, ".crawler")) // User config directory
		viper.AddConfigPath("/etc/crawler")                  // System config directory
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found - set up default configuration
		viper.Set("app.name", "gocrawl")
		viper.Set("app.environment", "development")
		viper.Set("app.debug", true)
		viper.Set("logger.level", "debug")
		viper.Set("logger.encoding", "console")
		viper.Set("logger.format", "text")
		viper.Set("logger.output", "stdout")
		viper.Set("crawler.source_file", "sources.yml")
		viper.Set("crawler.max_depth", defaultCrawlerMaxDepth)
		viper.Set("crawler.max_retries", defaultCrawlerMaxRetries)
		viper.Set("storage.batch_size", defaultStorageBatchSize)
		viper.Set("elasticsearch.max_retries", defaultElasticsearchRetries)
	}

	return nil
}

// bindFlags binds all command flags to viper configuration.
func bindFlags(cmd *cobra.Command) error {
	// Bind persistent flags
	if err := viper.BindPFlag("app.debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		return fmt.Errorf("failed to bind debug flag: %w", err)
	}

	if err := viper.BindPFlag("app.config_file", cmd.PersistentFlags().Lookup("config")); err != nil {
		return fmt.Errorf("failed to bind config flag: %w", err)
	}

	// Bind command-specific flags
	if cmd.Name() == "crawl" {
		if err := viper.BindPFlag("crawler.max_depth", cmd.Flags().Lookup("depth")); err != nil {
			return fmt.Errorf("failed to bind depth flag: %w", err)
		}
		if err := viper.BindPFlag("crawler.rate_limit", cmd.Flags().Lookup("rate-limit")); err != nil {
			return fmt.Errorf("failed to bind rate-limit flag: %w", err)
		}
	}

	return nil
}

// setupConfig handles configuration file setup for all commands
func setupConfig(cmd *cobra.Command) error {
	// Step 1: Set default values first
	setDefaults()

	// Step 2: Enable environment variable binding
	viper.AutomaticEnv()
	if err := bindEnvVars(); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Step 3: Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		return err
	}

	// Step 4: Load configuration file
	if err := loadConfigFile(); err != nil {
		return err
	}

	// Step 5: Bind command flags to viper
	if err := bindFlags(cmd); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	// Step 6: Validate configuration
	if err := validateConfig(cmd); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// validateConfig performs configuration validation based on the current command
func validateConfig(cmd *cobra.Command) error {
	// Validate app configuration
	if !viper.IsSet("app.name") {
		viper.Set("app.name", "gocrawl")
	}
	if !viper.IsSet("app.environment") {
		viper.Set("app.environment", "development")
	}

	// Command-specific validation
	switch cmd.Name() {
	case "crawl":
		if !viper.IsSet("crawler.base_url") {
			// Not required when using sources.yml
			viper.Set("crawler.base_url", "")
		}
		if !viper.IsSet("elasticsearch.url") {
			// Not required when using sources.yml
			viper.Set("elasticsearch.url", "http://localhost:9200")
		}
	case "search":
		if !viper.IsSet("elasticsearch.url") {
			viper.Set("elasticsearch.url", "http://localhost:9200")
		}
	}

	// Validate environment
	env := viper.GetString("app.environment")
	if env != "development" && env != "staging" && env != "production" {
		return fmt.Errorf("invalid environment: %s", env)
	}

	// Validate log level
	logLevel := viper.GetString("log.level")
	if logLevel != logLevelDebug && logLevel != logLevelInfo && logLevel != logLevelWarn && logLevel != logLevelError {
		viper.Set("log.level", logLevelInfo)
	}

	return nil
}

// Execute is the entry point for the CLI application.
// It runs the root command and handles any errors that occur during execution.
// If an error occurs, it prints the error message and exits with status code 1.
func Execute() {
	// Initialize configuration
	if err := setupConfig(rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize config
	cfg := config.NewConfig(log)
	if loadErr := cfg.Load(viper.GetString("config")); loadErr != nil {
		// If config file doesn't exist, use defaults
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(loadErr, &configFileNotFoundError) {
			log.Error("Failed to load config", "error", loadErr)
			os.Exit(1)
		}
	}

	// Create a context with dependencies
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.LoggerKey, log)
	ctx = context.WithValue(ctx, common.ConfigKey, cfg)
	rootCmd.SetContext(ctx)

	// Add commands first
	rootCmd.AddCommand(
		scheduler.Command(),         // Main scheduler command with fx
		index.Command(),             // For managing Elasticsearch index
		sources.NewSourcesCommand(), // For managing web content sources
		crawlcmd.Command(),          // For crawling web content
		httpdcmd.Command(),          // For running the HTTP server
		search.Command(),            // For searching content in Elasticsearch
	)

	if executeErr := rootCmd.Execute(); executeErr != nil {
		log.Error("Failed to execute command", "error", executeErr)
		os.Exit(1)
	}
}

// initLogger initializes a new logger instance
func initLogger() (logger.Interface, error) {
	// Get log level from config
	logLevel := viper.GetString("logger.level")
	level := logger.InfoLevel

	// Check debug flag from Viper first
	debug := viper.GetBool("app.debug")
	if debug {
		level = logger.DebugLevel
		logLevel = "debug"
	} else {
		// Fall back to log level from config
		switch logLevel {
		case "debug":
			level = logger.DebugLevel
		case "info":
			level = logger.InfoLevel
		case "warn":
			level = logger.WarnLevel
		case "error":
			level = logger.ErrorLevel
		}
	}

	// Create logger with configuration
	logConfig := &logger.Config{
		Level:       level,
		Development: debug,
		Encoding:    "console",
		EnableColor: true,
		OutputPaths: []string{"stdout"},
	}

	log, err := logger.New(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Log the current log level
	log.Debug("Root logger initialized",
		"level", logLevel,
		"debug", debug,
		"development", logConfig.Development)
	log.Info("Root logger ready",
		"level", logLevel,
		"debug", debug,
		"development", logConfig.Development)

	return log, nil
}

// init initializes the root command and its subcommands.
func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is ./config.yaml, ~/.crawler/config.yaml, or /etc/crawler/config.yaml)",
	)
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug mode")
}
