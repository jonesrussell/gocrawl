// Package config provides configuration management for the GoCrawl application.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

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
	TLS TLSConfig `yaml:"tls"`
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
		TLS TLSConfig `yaml:"tls"`
	}{
		Enabled:   viper.GetBool("server.security.enabled"),
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
		TLS: TLSConfig{
			Enabled:  viper.GetBool("server.security.tls.enabled"),
			CertFile: viper.GetString("server.security.tls.certificate"),
			KeyFile:  viper.GetString("server.security.tls.key"),
		},
	}
}
