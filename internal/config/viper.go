// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Default configuration values for local use only
const (
	defaultAppEnv     = "development"
	defaultAppName    = "gocrawl"
	defaultAppVersion = "1.0.0"
)

// viperMutex protects concurrent access to Viper operations
var viperMutex sync.Mutex

// setupViper initializes Viper with default configuration
func setupViper(log Logger) error {
	viperMutex.Lock()
	defer viperMutex.Unlock()

	// Configure Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add standard config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.gocrawl")
	viper.AddConfigPath("/etc/gocrawl")

	// Set environment variable prefix
	viper.SetEnvPrefix("GOCRAWL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	if err := bindEnvs(defaultEnvBindings()); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("failed to read config file",
			"error", err.Error(),
			"file", viper.ConfigFileUsed(),
		)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Set default values only if not already set
	if !viper.IsSet("app.environment") {
		viper.SetDefault("app.environment", defaultAppEnv)
	}
	if !viper.IsSet("app.name") {
		viper.SetDefault("app.name", defaultAppName)
	}
	if !viper.IsSet("app.version") {
		viper.SetDefault("app.version", defaultAppVersion)
	}
	if !viper.IsSet("app.debug") {
		viper.SetDefault("app.debug", false)
	}
	if !viper.IsSet("log.level") {
		viper.SetDefault("log.level", DefaultLogLevel)
	}
	if !viper.IsSet("log.debug") {
		viper.SetDefault("log.debug", false)
	}
	if !viper.IsSet("crawler.max_depth") {
		viper.SetDefault("crawler.max_depth", DefaultMaxDepth)
	}
	if !viper.IsSet("crawler.parallelism") {
		viper.SetDefault("crawler.parallelism", DefaultParallelism)
	}
	if !viper.IsSet("crawler.rate_limit") {
		viper.SetDefault("crawler.rate_limit", DefaultRateLimit)
	}
	if !viper.IsSet("crawler.random_delay") {
		viper.SetDefault("crawler.random_delay", DefaultRandomDelay)
	}
	if !viper.IsSet("server.host") {
		viper.SetDefault("server.host", DefaultHTTPHost)
	}
	if !viper.IsSet("server.port") {
		viper.SetDefault("server.port", DefaultServerPort)
	}
	if !viper.IsSet("server.read_timeout") {
		viper.SetDefault("server.read_timeout", DefaultHTTPReadTimeout)
	}
	if !viper.IsSet("server.write_timeout") {
		viper.SetDefault("server.write_timeout", DefaultHTTPWriteTimeout)
	}
	if !viper.IsSet("server.idle_timeout") {
		viper.SetDefault("server.idle_timeout", DefaultHTTPIdleTimeout)
	}
	if !viper.IsSet("storage.type") {
		viper.SetDefault("storage.type", DefaultStorageType)
	}
	if !viper.IsSet("storage.elasticsearch.host") {
		viper.SetDefault("storage.elasticsearch.host", DefaultElasticsearchHost)
	}
	if !viper.IsSet("storage.elasticsearch.index") {
		viper.SetDefault("storage.elasticsearch.index", DefaultElasticsearchIndex)
	}
	if !viper.IsSet("storage.elasticsearch.username") {
		viper.SetDefault("storage.elasticsearch.username", DefaultElasticsearchUsername)
	}
	if !viper.IsSet("storage.elasticsearch.password") {
		viper.SetDefault("storage.elasticsearch.password", DefaultElasticsearchPassword)
	}
	if !viper.IsSet("log.format") {
		viper.SetDefault("log.format", DefaultLogFormat)
	}
	if !viper.IsSet("log.output") {
		viper.SetDefault("log.output", DefaultLogOutput)
	}

	return nil
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

// defaultEnvBindings returns a map of viper config keys to environment variable names
func defaultEnvBindings() map[string]string {
	return map[string]string{
		"app.environment":               "APP_ENVIRONMENT",
		"app.name":                      "APP_NAME",
		"app.version":                   "APP_VERSION",
		"app.debug":                     "APP_DEBUG",
		"elasticsearch.username":        "ELASTIC_USERNAME",
		"elasticsearch.password":        "ELASTIC_PASSWORD",
		"elasticsearch.api_key":         "ELASTICSEARCH_API_KEY",
		"elasticsearch.tls.skip_verify": "ELASTIC_SKIP_TLS",
		"elasticsearch.tls.certificate": "ELASTIC_CERT_PATH",
		"elasticsearch.tls.key":         "ELASTIC_KEY_PATH",
		"elasticsearch.tls.ca":          "ELASTIC_CA_PATH",
		"server.address":                "GOCRAWL_PORT",
		"server.security.api_key":       "GOCRAWL_API_KEY",
		"log.level":                     "LOG_LEVEL",
		"log.debug":                     "LOG_DEBUG",
	}
}
