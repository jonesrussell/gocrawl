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
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/index"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

var (
	// cfgFile holds the path to the configuration file.
	// It can be set via the --config flag or defaults to config.yaml.
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

// Module provides the root command and its dependencies
var Module = fx.Module("root",
	common.Module,
	fx.Provide(
		func() (config.Interface, error) {
			return config.LoadConfig()
		},
		func(cfg config.Interface) (*zap.Logger, logger.Interface, error) {
			logConfig := &logger.Config{
				Level:       logger.Level(viper.GetString("logger.level")),
				Development: viper.GetBool("logger.development"),
				Encoding:    viper.GetString("logger.encoding"),
				OutputPaths: viper.GetStringSlice("logger.output_paths"),
				EnableColor: viper.GetBool("logger.enable_color"),
			}

			logInterface, zapLogger, err := logger.NewWithZap(logConfig)
			if err != nil {
				return nil, nil, err
			}

			return zapLogger, logInterface, nil
		},
	),
	fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log}
	}),
)

// Execute runs the root command
func Execute() error {
	// Initialize configuration
	if err := initConfig(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Bind flags
	if err := bindFlags(rootCmd); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create the application
	app := fx.New(
		Module,
		storage.ClientModule,
		index.Module,
		search.Module,
		httpd.Module,
	)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer app.Stop(context.Background())

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
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

	// Add index command
	rootCmd.AddCommand(index.Command)
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

	// Set defaults first
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, that's ok - we'll use defaults
		fmt.Fprintf(os.Stderr, "Warning: Config file not found: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Map environment variables to config keys
	viper.BindEnv("app.environment", "APP_ENV")
	viper.BindEnv("app.debug", "APP_DEBUG")
	viper.BindEnv("logger.level", "LOG_LEVEL")
	viper.BindEnv("logger.encoding", "LOG_FORMAT")
	viper.BindEnv("logger.development", "APP_DEBUG")  // Development mode follows debug mode
	viper.BindEnv("logger.enable_color", "APP_DEBUG") // Color output follows debug mode
	viper.BindEnv("logger.caller", "APP_DEBUG")       // Caller info follows debug mode
	viper.BindEnv("logger.stacktrace", "APP_DEBUG")   // Stacktrace follows debug mode

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file not found, that's ok - we'll use environment variables
		fmt.Fprintf(os.Stderr, "Warning: .env file not found: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Loaded .env file\n")
	}

	// Debug: Print all configuration values
	fmt.Fprintf(os.Stderr, "Configuration values:\n")
	fmt.Fprintf(os.Stderr, "  APP_ENV: %s\n", viper.GetString("app.environment"))
	fmt.Fprintf(os.Stderr, "  APP_DEBUG: %v\n", viper.GetBool("app.debug"))
	fmt.Fprintf(os.Stderr, "  LOG_LEVEL: %s\n", viper.GetString("logger.level"))
	fmt.Fprintf(os.Stderr, "  LOG_FORMAT: %s\n", viper.GetString("logger.encoding"))
	fmt.Fprintf(os.Stderr, "  LOG_DEVELOPMENT: %v\n", viper.GetBool("logger.development"))
	fmt.Fprintf(os.Stderr, "  LOG_ENABLE_COLOR: %v\n", viper.GetBool("logger.enable_color"))
	fmt.Fprintf(os.Stderr, "  LOG_CALLER: %v\n", viper.GetBool("logger.caller"))
	fmt.Fprintf(os.Stderr, "  LOG_STACKTRACE: %v\n", viper.GetBool("logger.stacktrace"))

	return nil
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
		"encoding":     "console",
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
	viper.SetDefault("elasticsearch", map[string]any{
		"addresses": []string{"http://localhost:9200"},
		"tls": map[string]any{
			"enabled":              true,
			"insecure_skip_verify": false,
		},
		"retry": map[string]any{
			"enabled":      true,
			"initial_wait": "1s",
			"max_wait":     "30s",
			"max_retries":  3,
		},
		"bulk_size":      1000,
		"flush_interval": "1s",
		"index_prefix":   "gocrawl",
		"discover_nodes": false,
	})

	// Crawler defaults - production safe
	viper.SetDefault("crawler", map[string]any{
		"max_depth":          3,
		"max_concurrency":    2,
		"request_timeout":    "30s",
		"user_agent":         "gocrawl/1.0",
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
		"parallelism": 2,
		"tls": map[string]any{
			"insecure_skip_verify": false,
		},
		"retry_delay":      "5s",
		"follow_redirects": true,
		"max_redirects":    5,
		"validate_urls":    true,
		"cleanup_interval": "1h",
	})
}
