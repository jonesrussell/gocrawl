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

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/index"
	cmdscheduler "github.com/jonesrussell/gocrawl/cmd/scheduler"
	"github.com/jonesrussell/gocrawl/cmd/search"
	cmdsources "github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}
			// Ensure the CommandContext is propagated from the root command to all subcommands
			// This is necessary because Cobra may create new contexts for subcommands
			rootCtx := cmd.Root().Context()
			if rootCtx != nil {
				if cmdCtx := rootCtx.Value(common.CommandContextKey); cmdCtx != nil {
					// If this command's context doesn't have the CommandContext, propagate it
					if cmd.Context().Value(common.CommandContextKey) == nil {
						newCtx := context.WithValue(cmd.Context(), common.CommandContextKey, cmdCtx)
						cmd.SetContext(newCtx)
					}
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

// Module provides the root command and its dependencies
var Module = fx.Module("root",
	common.Module,
	storage.ClientModule,
	storage.Module,
	crawl.Module,
	index.Module,
	cmdsources.Module,
	search.Module,
	httpd.Module,
	cmdscheduler.Module,
	fx.WithLogger(logger.NewFxLogger),
)

// Execute runs the root command
func Execute() error {
	// Load config and create logger
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logCfg := &logger.Config{
		Level:       logger.Level(viper.GetString("logger.level")),
		Development: viper.GetBool("logger.development"),
		Encoding:    viper.GetString("logger.encoding"),
		OutputPaths: viper.GetStringSlice("logger.output_paths"),
		EnableColor: viper.GetBool("logger.enable_color"),
	}

	log, err := logger.New(logCfg)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Create a context with dependencies using CommandContext
	cmdCtx := &common.CommandContext{
		Logger: log,
		Config: cfg,
	}
	ctx := context.WithValue(context.Background(), common.CommandContextKey, cmdCtx)

	// Initialize the application with the root module
	app := fx.New(
		Module,
		fx.Provide(
			func() config.Interface { return cfg },
			func() logger.Interface { return log },
		),
		fx.WithLogger(logger.NewFxLogger),
	)

	// Start the application
	if startErr := app.Start(ctx); startErr != nil {
		return fmt.Errorf("failed to start application: %w", startErr)
	}
	defer func() {
		if stopErr := app.Stop(ctx); stopErr != nil {
			log.Error("failed to stop application", "error", stopErr)
		}
	}()

	// Set the context on the root command to ensure it's available to all subcommands
	// This is necessary because ExecuteContext may not propagate the context to all commands
	rootCmd.SetContext(ctx)

	// Execute the root command with the context
	return rootCmd.ExecuteContext(ctx)
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
	rootCmd.AddCommand(index.Command)
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

	// Set defaults first
	setDefaults()

	// Read config file
	// Note: Config file is optional - if not found, we'll use defaults and environment variables
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, that's ok - we'll use defaults
		// This is expected behavior: config can come from file, environment variables, or defaults
		fmt.Fprintf(os.Stderr, "Warning: Config file not found: %v (using defaults and environment variables)\n", err)
	}

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Map environment variables to config keys
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

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file not found, that's ok - we'll use environment variables
		fmt.Fprintf(os.Stderr, "Warning: .env file not found: %v\n", err)
	}

	// Set development logging settings based on environment and debug flag
	isDev := viper.GetString("app.environment") == "development" || viper.GetBool("app.debug")
	if isDev {
		viper.Set("logger.development", true)
		viper.Set("logger.enable_color", true)
		viper.Set("logger.caller", true)
		viper.Set("logger.stacktrace", true)
		viper.Set("logger.encoding", "console")
		viper.Set("logger.level", "debug")
	}

	// Synchronize global Debug variable with Viper's value
	Debug = viper.GetBool("app.debug")

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
		"follow_redirects": true,
		"max_redirects":    crawler.DefaultMaxRedirects,
		"validate_urls":    true,
		"cleanup_interval": "1h",
	})
}
