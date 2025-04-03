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

	// Set default values
	viper.SetDefault("app.environment", defaultAppEnv)
	viper.SetDefault("app.name", defaultAppName)
	viper.SetDefault("app.version", defaultAppVersion)
	viper.SetDefault("app.debug", false)
	viper.SetDefault("log.level", DefaultLogLevel)
	viper.SetDefault("log.debug", false)
	viper.SetDefault("crawler.max_depth", defaultMaxDepth)
	viper.SetDefault("crawler.parallelism", defaultParallelism)
	viper.SetDefault("crawler.rate_limit", defaultCrawlerRateLimit)
	viper.SetDefault("crawler.random_delay", defaultRandomDelay)
	viper.SetDefault("elasticsearch.addresses", []string{defaultESAddress})
	viper.SetDefault("elasticsearch.index_name", defaultESIndex)
	viper.SetDefault("server.address", ":"+DefaultServerPort)
	viper.SetDefault("server.read_timeout", DefaultReadTimeout)
	viper.SetDefault("server.write_timeout", DefaultWriteTimeout)
	viper.SetDefault("server.idle_timeout", DefaultIdleTimeout)

	// Configure Viper
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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
		"elasticsearch.username":        "ELASTIC_USERNAME",
		"elasticsearch.password":        "ELASTIC_PASSWORD",
		"elasticsearch.api_key":         "ELASTICSEARCH_API_KEY",
		"elasticsearch.tls.skip_verify": "ELASTIC_SKIP_TLS",
		"elasticsearch.tls.certificate": "ELASTIC_CERT_PATH",
		"elasticsearch.tls.key":         "ELASTIC_KEY_PATH",
		"elasticsearch.tls.ca":          "ELASTIC_CA_PATH",
		"server.address":                "GOCRAWL_PORT",
		"server.security.api_key":       "GOCRAWL_API_KEY",
		"app.environment":               "APP_ENV",
		"app.debug":                     "APP_DEBUG",
		"log.level":                     "LOG_LEVEL",
		"log.debug":                     "LOG_DEBUG",
	}
}
