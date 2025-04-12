// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joho/godotenv"
	crawlcmd "github.com/jonesrussell/gocrawl/cmd/crawl"
	httpdcmd "github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/cmd/sources"
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

	// log is the global logger instance
	log logger.Interface
)

// bindEnvVars binds environment variables to configuration keys
func bindEnvVars() error {
	// App configuration
	if err := viper.BindEnv("app.name", "APP_NAME"); err != nil {
		return fmt.Errorf("failed to bind app.name: %w", err)
	}
	if err := viper.BindEnv("app.environment", "APP_ENV"); err != nil {
		return fmt.Errorf("failed to bind app.environment: %w", err)
	}
	if err := viper.BindEnv("app.debug", "APP_DEBUG"); err != nil {
		return fmt.Errorf("failed to bind app.debug: %w", err)
	}

	// Server configuration
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

	// Elasticsearch configuration
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
	if err := viper.BindEnv("elasticsearch.tls_insecure_skip_verify", "ELASTICSEARCH_TLS_INSECURE_SKIP_VERIFY"); err != nil {
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

	// Crawler configuration
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

	// Logging configuration
	if err := viper.BindEnv("logger.level", "LOG_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind logger.level: %w", err)
	}
	if err := viper.BindEnv("logger.format", "LOG_FORMAT"); err != nil {
		return fmt.Errorf("failed to bind logger.format: %w", err)
	}

	// Server Security Configuration
	if err := viper.BindEnv("server.port", "GOCRAWL_PORT"); err != nil {
		return fmt.Errorf("failed to bind server.port: %w", err)
	}
	if err := viper.BindEnv("server.api_key", "GOCRAWL_API_KEY"); err != nil {
		return fmt.Errorf("failed to bind server.api_key: %w", err)
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
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "console")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.enable_color", true)

	// Crawler defaults
	viper.SetDefault("crawler.max_depth", 10)
	viper.SetDefault("crawler.max_retries", 3)
	viper.SetDefault("crawler.rate_limit", "1s")
	viper.SetDefault("crawler.timeout", "30s")
	viper.SetDefault("crawler.user_agent", "GoCrawl/1.0")

	// Storage defaults
	viper.SetDefault("storage.type", "elasticsearch")
	viper.SetDefault("storage.batch_size", 100)
	viper.SetDefault("storage.flush_interval", "5s")

	// Elasticsearch defaults
	viper.SetDefault("elasticsearch.url", "https://localhost:9200")
	viper.SetDefault("elasticsearch.sniff", false)
	viper.SetDefault("elasticsearch.healthcheck", true)
	viper.SetDefault("elasticsearch.retry_on_conflict", 3)
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

// setupConfig handles configuration file setup for all commands.
// It ensures the config file path is absolute and configures Viper.
func setupConfig(cmd *cobra.Command) error {
	// Step 1: Set default values first
	setDefaults()

	// Step 2: Enable environment variable binding
	viper.AutomaticEnv()
	if err := bindEnvVars(); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Step 3: Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		// .env file not found is not an error - we'll use environment variables
	}

	// Step 4: Get the absolute path to the config file if specified
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

	// Step 5: Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is not an error - we'll use defaults
	}

	// Step 6: Bind command flags to viper
	if err := bindFlags(cmd); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	// Step 7: Validate configuration
	if err := validateConfig(cmd); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// validateConfig performs configuration validation based on the current command.
func validateConfig(cmd *cobra.Command) error {
	// Validate app configuration
	if !viper.IsSet("app.name") {
		return errors.New("app.name is required")
	}
	if !viper.IsSet("app.environment") {
		return errors.New("app.environment is required")
	}

	// Command-specific validation
	switch cmd.Name() {
	case "crawl":
		if !viper.IsSet("crawler.base_url") {
			return errors.New("crawler.base_url is required for crawl command")
		}
		if !viper.IsSet("elasticsearch.url") {
			return errors.New("elasticsearch.url is required for crawl command")
		}
	case "search":
		if !viper.IsSet("elasticsearch.url") {
			return errors.New("elasticsearch.url is required for search command")
		}
	}

	// Validate environment
	env := viper.GetString("app.environment")
	if env != "development" && env != "staging" && env != "production" {
		return fmt.Errorf("invalid environment: %s", env)
	}

	// Validate log level
	logLevel := viper.GetString("log.level")
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" {
		return fmt.Errorf("invalid log level: %s", logLevel)
	}

	return nil
}

// initLogger initializes the global logger instance
func initLogger() error {
	// Get log level from config
	logLevel := viper.GetString("logger.level")
	level := logger.InfoLevel
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

	// Override with debug flag if set
	if Debug {
		level = logger.DebugLevel
		fmt.Fprintf(os.Stderr, "Debug mode enabled, setting log level to DEBUG\n")
	}

	// Create logger with configuration
	logConfig := &logger.Config{
		Level:       level,
		Development: Debug,
		Encoding:    viper.GetString("logger.format"),
		EnableColor: true, // Always enable color in debug mode
		OutputPaths: []string{viper.GetString("logger.output")},
	}

	var err error
	log, err = logger.New(logConfig)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	return nil
}

// Execute is the entry point for the CLI application.
// It runs the root command and handles any errors that occur during execution.
// If an error occurs, it prints the error message and exits with status code 1.
func Execute() {
	// Initialize logger first
	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create a context with the logger
	ctx := context.WithValue(context.Background(), common.LoggerKey, log)
	rootCmd.SetContext(ctx)

	// Add commands
	rootCmd.AddCommand(
		job.NewJobCommand(log),         // Main job command with fx
		job.NewJobSubCommands(log),     // Job subcommands
		indices.Command(),              // For managing Elasticsearch indices
		sources.NewSourcesCommand(log), // For managing web content sources
		crawlcmd.Command(),             // For crawling web content
		httpdcmd.Command(),             // For running the HTTP server
		search.Command(),               // For searching content in Elasticsearch
	)

	if executeErr := rootCmd.Execute(); executeErr != nil {
		log.Error("Failed to execute command", "error", executeErr)
		os.Exit(1)
	}
}

// init initializes the root command and its subcommands.
func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml, ~/.crawler/config.yaml, or /etc/crawler/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug mode")

	// Initialize configuration on startup
	cobra.OnInitialize(func() {
		if err := setupConfig(rootCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
	})

	// Remove the PersistentPreRunE to avoid double loading
	rootCmd.PersistentPreRunE = nil
}
