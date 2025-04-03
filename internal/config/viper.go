// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// viperMutex protects concurrent access to Viper operations
var viperMutex sync.Mutex

// setupViper initializes Viper with default configuration
func setupViper(log Logger) error {
	viperMutex.Lock()
	defer viperMutex.Unlock()

	// Configure Viper
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	if err := bindEnvs(defaultEnvBindings()); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Read config file if specified
	if configFile := viper.GetString("config_file"); configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			if !os.IsNotExist(err) {
				log.Warn("Error reading config file",
					String("file", configFile),
					Error(err))
				return fmt.Errorf("failed to read config file: %w", err)
			}
		}
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
		viper.SetDefault("crawler.max_depth", defaultMaxDepth)
	}
	if !viper.IsSet("crawler.parallelism") {
		viper.SetDefault("crawler.parallelism", defaultParallelism)
	}
	if !viper.IsSet("crawler.rate_limit") {
		viper.SetDefault("crawler.rate_limit", defaultCrawlerRateLimit)
	}
	if !viper.IsSet("crawler.random_delay") {
		viper.SetDefault("crawler.random_delay", defaultRandomDelay)
	}
	if !viper.IsSet("elasticsearch.addresses") {
		viper.SetDefault("elasticsearch.addresses", []string{defaultESAddress})
	}
	if !viper.IsSet("elasticsearch.index_name") {
		viper.SetDefault("elasticsearch.index_name", defaultESIndex)
	}
	if !viper.IsSet("server.address") {
		viper.SetDefault("server.address", ":"+DefaultServerPort)
	}
	if !viper.IsSet("server.read_timeout") {
		viper.SetDefault("server.read_timeout", DefaultReadTimeout)
	}
	if !viper.IsSet("server.write_timeout") {
		viper.SetDefault("server.write_timeout", DefaultWriteTimeout)
	}
	if !viper.IsSet("server.idle_timeout") {
		viper.SetDefault("server.idle_timeout", DefaultIdleTimeout)
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
