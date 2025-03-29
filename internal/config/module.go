// Package config provides configuration management for the GoCrawl application.
// This file specifically handles dependency injection and module initialization
// using the fx framework.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	// defaultRetryMaxWait is the default maximum wait time between retries
	defaultRetryMaxWait = 30 * time.Second

	// defaultRetryInitialWait is the default initial wait time between retries
	defaultRetryInitialWait = 1 * time.Second

	// defaultMaxRetries is the default number of retries for failed requests
	defaultMaxRetries = 3

	// DefaultReadTimeout Server timeouts
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultIdleTimeout  = 60 * time.Second
	DefaultServerPort   = "8080"

	// Environment types
	envProduction = "production"

	// Default crawler settings
	defaultMaxDepth    = 3
	defaultParallelism = 5

	// Constants for default configuration values
	defaultMaxAge             = 86400 // 24 hours in seconds
	defaultRateLimitPerMinute = 60
)

// Impl implements the Interface and holds the actual configuration values
type Impl struct {
	App           AppConfig           `yaml:"app"`
	Crawler       CrawlerConfig       `yaml:"crawler"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Log           LogConfig           `yaml:"log"`
	Sources       []Source            `yaml:"sources"`
	Server        ServerConfig        `yaml:"server"`
	Command       string              `yaml:"command"`
}

// Ensure Impl implements Interface
var _ Interface = (*Impl)(nil)

// GetCrawlerConfig implements Interface
func (c *Impl) GetCrawlerConfig() *CrawlerConfig {
	return &c.Crawler
}

// GetElasticsearchConfig implements Interface
func (c *Impl) GetElasticsearchConfig() *ElasticsearchConfig {
	return &c.Elasticsearch
}

// GetLogConfig implements Interface
func (c *Impl) GetLogConfig() *LogConfig {
	return &c.Log
}

// GetAppConfig implements Interface
func (c *Impl) GetAppConfig() *AppConfig {
	return &c.App
}

// GetSources implements Interface
func (c *Impl) GetSources() []Source {
	return c.Sources
}

// GetServerConfig implements Interface
func (c *Impl) GetServerConfig() *ServerConfig {
	return &c.Server
}

// GetCommand implements Interface
func (c *Impl) GetCommand() string {
	return c.Command
}

// bindEnvs binds environment variables to their viper config keys
func bindEnvs(bindings map[string]string) error {
	for k, v := range bindings {
		if err := viper.BindEnv(k, v); err != nil {
			return fmt.Errorf("failed to bind env var %s: %w", v, err)
		}
	}
	return nil
}

// loadEnvFile loads an environment file if it exists and logs any errors
func loadEnvFile(envFile string) error {
	if err := godotenv.Load(envFile); err != nil {
		if !os.IsNotExist(err) {
			defaultLog.Warn("Error loading environment file", "file", envFile, "error", err)
			return err
		}
		return nil
	}
	defaultLog.Info("Loaded environment file", "file", envFile)
	return nil
}

// SetupConfig initializes the configuration system with an optional environment file
func SetupConfig(envFile string) error {
	// Initialize Viper configuration
	if err := setupViper(); err != nil {
		return fmt.Errorf("failed to setup Viper: %w", err)
	}

	// Load environment file
	if envFile != "" {
		if err := loadEnvFile(envFile); err != nil {
			return fmt.Errorf("failed to load environment file: %w", err)
		}
	} else {
		// Load default .env file if it exists
		if err := loadEnvFile(".env"); err != nil {
			return fmt.Errorf("failed to load default .env file: %w", err)
		}
	}

	// Bind environment variables
	if err := bindEnvs(defaultEnvBindings()); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	return nil
}

// Module provides the config module and its dependencies using fx.
// It sets up the configuration providers that can be used throughout
// the application for dependency injection.
//
// The module provides:
// - Interface instance via the New constructor
// - HTTP transport configuration via NewHTTPTransport
var Module = fx.Options(
	fx.Provide(
		// Provide the config interface
		func() (Interface, error) {
			if err := SetupConfig(""); err != nil {
				return nil, err
			}
			return New()
		},
		NewHTTPTransport, // Provides HTTP transport configuration
	),
)

// TestModule provides a test configuration module that loads the test environment file
var TestModule = fx.Options(
	fx.Provide(
		// Provide the config interface for tests
		func() (Interface, error) {
			if err := SetupConfig(".env.test"); err != nil {
				return nil, err
			}
			return New()
		},
		NewHTTPTransport,
	),
)

// defaultLog implements the logger interface using standard logging
var defaultLog = struct {
	Info func(msg string, args ...any)
	Warn func(msg string, args ...any)
}{
	Info: func(msg string, args ...any) {
		if len(args) > 0 {
			// Format key-value pairs
			formattedArgs := make([]any, 0, len(args))
			for i := 0; i < len(args); i += 2 {
				if i+1 < len(args) {
					formattedArgs = append(formattedArgs, args[i], args[i+1])
				}
			}
			_, err := os.Stdout.WriteString(fmt.Sprintf("INFO: %s %v\n", msg, formattedArgs))
			if err != nil {
				return
			}
		} else {
			_, err := os.Stdout.WriteString(fmt.Sprintf("INFO: %s\n", msg))
			if err != nil {
				return
			}
		}
	},
	Warn: func(msg string, args ...any) {
		if len(args) > 0 {
			// Format key-value pairs
			formattedArgs := make([]any, 0, len(args))
			for i := 0; i < len(args); i += 2 {
				if i+1 < len(args) {
					formattedArgs = append(formattedArgs, args[i], args[i+1])
				}
			}
			_, err := os.Stderr.WriteString(fmt.Sprintf("WARN: %s %v\n", msg, formattedArgs))
			if err != nil {
				return
			}
		} else {
			_, err := os.Stderr.WriteString(fmt.Sprintf("WARN: %s\n", msg))
			if err != nil {
				return
			}
		}
	},
}

// setupViper initializes Viper with default configuration
func setupViper() error {
	// Load config file from environment if specified
	if cfgFile := os.Getenv("GOCRAWL_CONFIG"); cfgFile != "" {
		defaultLog.Info("Using config file from environment", "file", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config file in current directory
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.gocrawl")
		viper.AddConfigPath("/etc/gocrawl")
	}

	// Read config file first
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("error reading config file: %w", err)
		}
		defaultLog.Warn("No config file found, using defaults")
	} else {
		defaultLog.Info("Configuration loaded from", "file", viper.ConfigFileUsed())
	}

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			defaultLog.Warn("Error loading .env file", "error", err)
		}
	}

	// Configure environment variable handling last
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values only if not already set
	if !viper.IsSet("app.environment") {
		viper.SetDefault("app.environment", "development")
	}
	if !viper.IsSet("log.level") {
		viper.SetDefault("log.level", "info")
	}
	if !viper.IsSet("crawler.source_file") {
		viper.SetDefault("crawler.source_file", "sources.yml")
	}
	if !viper.IsSet("elasticsearch.addresses") {
		viper.SetDefault("elasticsearch.addresses", []string{"https://localhost:9200"})
	}
	if !viper.IsSet("crawler.max_depth") {
		viper.SetDefault("crawler.max_depth", defaultMaxDepth)
	}
	if !viper.IsSet("crawler.parallelism") {
		viper.SetDefault("crawler.parallelism", defaultParallelism)
	}

	return nil
}

// defaultEnvBindings returns the default environment variable bindings
func defaultEnvBindings() map[string]string {
	return map[string]string{
		"elasticsearch.username":        "ELASTIC_USERNAME",
		"elasticsearch.password":        "ELASTIC_PASSWORD",
		"elasticsearch.api_key":         "ELASTIC_API_KEY",
		"elasticsearch.tls.skip_verify": "ELASTIC_SKIP_TLS",
		"elasticsearch.tls.certificate": "ELASTIC_CERT_PATH",
		"elasticsearch.tls.key":         "ELASTIC_KEY_PATH",
		"elasticsearch.tls.ca":          "ELASTIC_CA_PATH",
		"server.address":                "GOCRAWL_PORT",
		"server.security.api_key":       "GOCRAWL_API_KEY",
		"app.environment":               "APP_ENV",
		"log.level":                     "LOG_LEVEL",
	}
}

// createServerSecurityConfig creates the server security configuration
func createServerSecurityConfig() struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key"`
	RateLimit int    `yaml:"rate_limit"`
	CORS      struct {
		Enabled        bool     `yaml:"enabled"`
		AllowedOrigins []string `yaml:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers"`
		MaxAge         int      `yaml:"max_age"`
	} `yaml:"cors"`
	TLS struct {
		Enabled     bool   `yaml:"enabled"`
		Certificate string `yaml:"certificate"`
		Key         string `yaml:"key"`
	} `yaml:"tls"`
} {
	return struct {
		Enabled   bool   `yaml:"enabled"`
		APIKey    string `yaml:"api_key"`
		RateLimit int    `yaml:"rate_limit"`
		CORS      struct {
			Enabled        bool     `yaml:"enabled"`
			AllowedOrigins []string `yaml:"allowed_origins"`
			AllowedMethods []string `yaml:"allowed_methods"`
			AllowedHeaders []string `yaml:"allowed_headers"`
			MaxAge         int      `yaml:"max_age"`
		} `yaml:"cors"`
		TLS struct {
			Enabled     bool   `yaml:"enabled"`
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
		} `yaml:"tls"`
	}{
		Enabled:   true,
		APIKey:    viper.GetString("server.security.api_key"),
		RateLimit: defaultRateLimitPerMinute,
		CORS: struct {
			Enabled        bool     `yaml:"enabled"`
			AllowedOrigins []string `yaml:"allowed_origins"`
			AllowedMethods []string `yaml:"allowed_methods"`
			AllowedHeaders []string `yaml:"allowed_headers"`
			MaxAge         int      `yaml:"max_age"`
		}{
			Enabled:        true,
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-API-Key"},
			MaxAge:         defaultMaxAge,
		},
	}
}

// createElasticsearchConfig creates the Elasticsearch configuration
func createElasticsearchConfig() ElasticsearchConfig {
	// Get retry settings with defaults
	retryEnabled := viper.GetBool("elasticsearch.retry.enabled")
	retryInitialWait := viper.GetDuration("elasticsearch.retry.initial_wait")
	if retryInitialWait == 0 {
		retryInitialWait = defaultRetryInitialWait
	}
	retryMaxWait := viper.GetDuration("elasticsearch.retry.max_wait")
	if retryMaxWait == 0 {
		retryMaxWait = defaultRetryMaxWait
	}
	retryMaxRetries := viper.GetInt("elasticsearch.retry.max_retries")
	if retryMaxRetries == 0 {
		retryMaxRetries = defaultMaxRetries
	}

	return ElasticsearchConfig{
		Addresses: viper.GetStringSlice("elasticsearch.addresses"),
		Username:  viper.GetString("elasticsearch.username"),
		Password:  viper.GetString("elasticsearch.password"),
		APIKey:    viper.GetString("elasticsearch.api_key"),
		IndexName: viper.GetString("elasticsearch.index_name"),
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{
			ID:     viper.GetString("elasticsearch.cloud.id"),
			APIKey: viper.GetString("elasticsearch.cloud.api_key"),
		},
		TLS: struct {
			Enabled     bool   `yaml:"enabled"`
			SkipVerify  bool   `yaml:"skip_verify"`
			Certificate string `yaml:"certificate"`
			Key         string `yaml:"key"`
			CA          string `yaml:"ca"`
		}{
			Enabled:     viper.GetBool("elasticsearch.tls.enabled"),
			SkipVerify:  viper.GetBool("elasticsearch.tls.skip_verify"),
			Certificate: viper.GetString("elasticsearch.tls.certificate"),
			Key:         viper.GetString("elasticsearch.tls.key"),
			CA:          viper.GetString("elasticsearch.tls.ca"),
		},
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     retryEnabled,
			InitialWait: retryInitialWait,
			MaxWait:     retryMaxWait,
			MaxRetries:  retryMaxRetries,
		},
	}
}

// createCrawlerConfig creates the crawler configuration
func createCrawlerConfig() (CrawlerConfig, error) {
	rateLimit, err := parseRateLimit(viper.GetString("crawler.rate_limit"))
	if err != nil {
		return CrawlerConfig{}, fmt.Errorf("error parsing rate limit: %w", err)
	}

	return CrawlerConfig{
		BaseURL:          viper.GetString("crawler.base_url"),
		MaxDepth:         viper.GetInt("crawler.max_depth"),
		RateLimit:        rateLimit,
		RandomDelay:      viper.GetDuration("crawler.random_delay"),
		IndexName:        viper.GetString("crawler.index_name"),
		ContentIndexName: viper.GetString("crawler.content_index_name"),
		SourceFile:       viper.GetString("crawler.source_file"),
		Parallelism:      viper.GetInt("crawler.parallelism"),
	}, nil
}

// createServerConfig creates the server configuration
func createServerConfig() ServerConfig {
	// Get server timeouts with defaults
	readTimeout := viper.GetDuration("server.read_timeout")
	if readTimeout == 0 {
		readTimeout = DefaultReadTimeout
	}
	writeTimeout := viper.GetDuration("server.write_timeout")
	if writeTimeout == 0 {
		writeTimeout = DefaultWriteTimeout
	}
	idleTimeout := viper.GetDuration("server.idle_timeout")
	if idleTimeout == 0 {
		idleTimeout = DefaultIdleTimeout
	}

	// Get server address with default and ensure proper format
	address := viper.GetString("server.address")
	if address == "" {
		address = ":" + DefaultServerPort
	} else if !strings.Contains(address, ":") {
		address = ":" + address
	}

	return ServerConfig{
		Address:      address,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Security:     createServerSecurityConfig(),
	}
}

// createConfig creates a new config instance from Viper settings
func createConfig() (*Impl, error) {
	// Get the command being run from os.Args
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Create crawler config
	crawlerConfig, err := createCrawlerConfig()
	if err != nil {
		return nil, err
	}

	// Load sources from the source file
	sourcesConfig, err := loader.LoadFromFile(crawlerConfig.SourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources from %s: %w", crawlerConfig.SourceFile, err)
	}

	// Convert loader.Config to config.Source
	var sources []Source
	for _, src := range sourcesConfig {
		rateLimit, parseErr := ParseRateLimit(src.RateLimit)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse rate limit for source %s: %w", src.Name, parseErr)
		}

		sources = append(sources, Source{
			Name:         src.Name,
			URL:          src.URL,
			RateLimit:    rateLimit,
			MaxDepth:     src.MaxDepth,
			ArticleIndex: src.ArticleIndex,
			Index:        src.Index,
			Time:         src.Time,
			Selectors: SourceSelectors{
				Article: ArticleSelectors{
					Container:     src.Selectors.Article.Container,
					Title:         src.Selectors.Article.Title,
					Body:          src.Selectors.Article.Body,
					Intro:         src.Selectors.Article.Intro,
					Byline:        src.Selectors.Article.Byline,
					PublishedTime: src.Selectors.Article.PublishedTime,
					TimeAgo:       src.Selectors.Article.TimeAgo,
					JSONLD:        src.Selectors.Article.JSONLD,
					Section:       src.Selectors.Article.Section,
					Keywords:      src.Selectors.Article.Keywords,
					Description:   src.Selectors.Article.Description,
					OGTitle:       src.Selectors.Article.OGTitle,
					OGDescription: src.Selectors.Article.OGDescription,
					OGImage:       src.Selectors.Article.OGImage,
					OgURL:         src.Selectors.Article.OgURL,
					Canonical:     src.Selectors.Article.Canonical,
					WordCount:     src.Selectors.Article.WordCount,
					PublishDate:   src.Selectors.Article.PublishDate,
					Category:      src.Selectors.Article.Category,
					Tags:          src.Selectors.Article.Tags,
					Author:        src.Selectors.Article.Author,
					BylineName:    src.Selectors.Article.BylineName,
				},
			},
		})
	}

	// Create the config struct
	cfg := &Impl{
		App: AppConfig{
			Environment: viper.GetString("app.environment"),
			Name:        viper.GetString("app.name"),
			Version:     viper.GetString("app.version"),
			Debug:       viper.GetBool("app.debug"),
		},
		Crawler:       crawlerConfig,
		Elasticsearch: createElasticsearchConfig(),
		Log: LogConfig{
			Level: viper.GetString("log.level"),
			Debug: viper.GetBool("log.debug"),
		},
		Server:  createServerConfig(),
		Command: command,
		Sources: sources,
	}

	// Validate the configuration
	if validationErr := ValidateConfig(cfg); validationErr != nil {
		return nil, fmt.Errorf("invalid configuration: %w", validationErr)
	}

	return cfg, nil
}

// New creates a new config provider
func New() (*Impl, error) {
	// Load .env file in development mode
	if os.Getenv("APP_ENV") != envProduction {
		if loadErr := godotenv.Load(); loadErr != nil {
			// Only log a warning as .env file is optional
			defaultLog.Warn("Error loading .env file", "error", loadErr)
		}
	}

	// Create configuration from Viper settings
	cfg, err := createConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return cfg, nil
}
