// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/index"
	cmdscheduler "github.com/jonesrussell/gocrawl/cmd/scheduler"
	"github.com/jonesrussell/gocrawl/cmd/search"
	cmdsources "github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
)

var (
	// cfgFile holds the path to the configuration file.
	cfgFile string

	// Debug enables debug mode for all commands
	Debug bool

	// rootCmd represents the root command for the GoCrawl CLI.
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler and search engine",
		Long:  `A web crawler and search engine built with Go.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

// Execute runs the root command
func Execute() error {
	// Load .env file early so environment variables are available
	_ = godotenv.Load()

	// Parse flags early to get debug flag before creating logger
	_ = rootCmd.ParseFlags(os.Args[1:])

	// Initialize configuration
	if err := initConfig(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Execute the root command with a fresh context
	return rootCmd.ExecuteContext(context.Background())
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

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gocrawl version %s\n", "1.0.0") // TODO: Get from build info
		},
	})

	// Add subcommands
	rootCmd.AddCommand(crawl.Command())
	rootCmd.AddCommand(index.Command())
	rootCmd.AddCommand(cmdsources.NewSourcesCommand())
	rootCmd.AddCommand(search.Command())
	rootCmd.AddCommand(httpd.Command())
	rootCmd.AddCommand(cmdscheduler.Command())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// Set config file
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	// Load .env file first, before setting defaults and reading config
	// This ensures environment variables from .env are available when Viper reads them
	// Note: This may be called twice (once in Execute(), once here), but godotenv.Load()
	// is idempotent and won't overwrite existing environment variables
	if err := godotenv.Load(); err != nil {
		// .env file not found, that's ok - we'll use environment variables
		fmt.Fprintf(os.Stderr, "Warning: .env file not found: %v\n", err)
	}

	// Enable automatic environment variable reading BEFORE setting defaults
	// This ensures environment variables take precedence over defaults
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults (only used if environment variables or config file don't provide values)
	setDefaults()

	// Read config file
	// Note: Config file is optional - if not found, we'll use defaults and environment variables
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, that's ok - we'll use defaults
		// This is expected behavior: config can come from file, environment variables, or defaults
		fmt.Fprintf(os.Stderr, "Warning: Config file not found: %v (using defaults and environment variables)\n", err)
	}

	// Bind command-line flags to Viper
	if err := bindCommandLineFlags(); err != nil {
		return err
	}

	// Map environment variables to config keys
	if err := bindAppEnvVars(); err != nil {
		return err
	}

	// Bind Elasticsearch environment variables
	if err := bindElasticsearchEnvVars(); err != nil {
		return err
	}

	// Set development logging settings
	setupDevelopmentLogging()

	return nil
}

// bindCommandLineFlags binds command-line flags to Viper.
func bindCommandLineFlags() error {
	if err := viper.BindPFlag("app.debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		return fmt.Errorf("failed to bind debug flag: %w", err)
	}
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		return fmt.Errorf("failed to bind config flag: %w", err)
	}
	return nil
}

// bindAppEnvVars binds application and logger environment variables to config keys.
func bindAppEnvVars() error {
	if err := viper.BindEnv("app.environment", "APP_ENV"); err != nil {
		return fmt.Errorf("failed to bind APP_ENV: %w", err)
	}
	if err := viper.BindEnv("app.debug", "APP_DEBUG"); err != nil {
		return fmt.Errorf("failed to bind APP_DEBUG: %w", err)
	}
	if err := viper.BindEnv("logger.level", "LOG_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind LOG_LEVEL: %w", err)
	}
	if err := viper.BindEnv("logger.encoding", "LOG_FORMAT"); err != nil {
		return fmt.Errorf("failed to bind LOG_FORMAT: %w", err)
	}
	return nil
}

// bindElasticsearchEnvVars binds Elasticsearch environment variables to config keys.
func bindElasticsearchEnvVars() error {
	// Support both ELASTICSEARCH_HOSTS and ELASTICSEARCH_ADDRESSES
	// Note: ELASTICSEARCH_ADDRESSES is also handled by AutomaticEnv via the replacer,
	// but we explicitly bind ELASTICSEARCH_HOSTS for clarity
	if err := viper.BindEnv("elasticsearch.addresses", "ELASTICSEARCH_HOSTS", "ELASTICSEARCH_ADDRESSES"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch addresses: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.password", "ELASTIC_PASSWORD", "ELASTICSEARCH_PASSWORD"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch password: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.tls.insecure_skip_verify", "ELASTICSEARCH_SKIP_TLS"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch TLS skip verify: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.api_key", "ELASTICSEARCH_API_KEY"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch API key: %w", err)
	}
	// Bind index_name (supports both ELASTICSEARCH_INDEX_PREFIX and ELASTICSEARCH_INDEX_NAME)
	if err := viper.BindEnv("elasticsearch.index_name",
		"ELASTICSEARCH_INDEX_PREFIX", "ELASTICSEARCH_INDEX_NAME"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch index name: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.retry.max_retries", "ELASTICSEARCH_MAX_RETRIES"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch max retries: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.retry.initial_wait", "ELASTICSEARCH_RETRY_INITIAL_WAIT"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch retry initial wait: %w", err)
	}
	if err := viper.BindEnv("elasticsearch.retry.max_wait", "ELASTICSEARCH_RETRY_MAX_WAIT"); err != nil {
		return fmt.Errorf("failed to bind Elasticsearch retry max wait: %w", err)
	}
	return nil
}

// setupDevelopmentLogging configures development logging settings based on environment and debug flag.
func setupDevelopmentLogging() {
	// Check both the flag variable and Viper to ensure we catch the debug flag
	// Note: Debug variable is set by ParseFlags(), and we bind it to Viper above
	debugFlag := Debug || viper.GetBool("app.debug")
	isDev := viper.GetString("app.environment") == "development"

	// Only set debug level if explicitly requested via flag or APP_DEBUG
	// Do NOT automatically set debug level just because environment is "development"
	if debugFlag {
		viper.Set("logger.level", "debug")
	}

	// Set development mode features (formatting, colors, etc.) if in development environment
	// But do NOT change log level unless explicitly requested
	if isDev {
		viper.Set("logger.development", true)
		viper.Set("logger.enable_color", true)
		viper.Set("logger.caller", true)
		viper.Set("logger.stacktrace", true)
		viper.Set("logger.encoding", "console")
		// Only set debug level if explicitly requested
		if debugFlag {
			viper.Set("logger.level", "debug")
		}
	}

	// Synchronize global Debug variable with Viper's value
	Debug = debugFlag
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults - production safe
	viper.SetDefault("app", map[string]any{
		"name":        "gocrawl",
		"version":     "1.0.0",
		"environment": "production",
		"debug":       false,
	})

	// Logger defaults - production safe
	viper.SetDefault("logger", map[string]any{
		"level":        "info",
		"development":  false,
		"encoding":     "json",
		"output_paths": []string{"stdout"},
		"enable_color": false,
		"caller":       false,
		"stacktrace":   false,
		"max_size":     config.DefaultMaxLogSize,
		"max_backups":  config.DefaultMaxLogBackups,
		"max_age":      config.DefaultMaxLogAge,
		"compress":     true,
	})

	// Server defaults - production safe
	viper.SetDefault("server", map[string]any{
		"address":          ":8080",
		"read_timeout":     "15s",
		"write_timeout":    "15s",
		"idle_timeout":     "60s",
		"security_enabled": true,
	})

	// Elasticsearch defaults - production safe
	// Use 127.0.0.1 instead of localhost to avoid IPv6 resolution issues
	viper.SetDefault("elasticsearch", map[string]any{
		"addresses": []string{"http://127.0.0.1:9200"},
		"tls": map[string]any{
			"enabled":              true,
			"insecure_skip_verify": false,
		},
		"retry": map[string]any{
			"enabled":      true,
			"initial_wait": "1s",
			"max_wait":     "30s",
			"max_retries":  crawler.DefaultMaxRetries,
		},
		"bulk_size":      config.DefaultBulkSize,
		"flush_interval": "1s",
		"index_prefix":   "gocrawl",
		"discover_nodes": false,
	})

	// Crawler defaults - production safe
	viper.SetDefault("crawler", map[string]any{
		"max_depth":          crawler.DefaultMaxDepth,
		"max_concurrency":    crawler.DefaultParallelism,
		"request_timeout":    "30s",
		"user_agent":         crawler.DefaultUserAgent,
		"respect_robots_txt": true,
		"delay":              "1s",
		"random_delay":       "0s",
		"source_file":        "sources.yml",
		"debugger": map[string]any{
			"enabled": false,
			"level":   "info",
			"format":  "json",
			"output":  "stdout",
		},
		"rate_limit":  "2s",
		"parallelism": crawler.DefaultParallelism,
		"tls": map[string]any{
			"insecure_skip_verify": false,
		},
		"retry_delay":      "5s",
		"max_retries":      crawler.DefaultMaxRetries,
		"follow_redirects": true,
		"max_redirects":    crawler.DefaultMaxRedirects,
		"validate_urls":    true,
		"cleanup_interval": crawler.DefaultCleanupInterval.String(),
	})
}
