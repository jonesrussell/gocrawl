// Package storage provides storage-related configuration types and functions.
package storage

import (
	"errors"
	"fmt"
)

// Default configuration values
const (
	DefaultType     = "elasticsearch"
	DefaultHost     = "localhost"
	DefaultPort     = 9200
	DefaultUsername = ""
	DefaultPassword = ""
	DefaultSSL      = false
)

// Config holds storage-specific configuration settings.
type Config struct {
	// Type is the storage type (elasticsearch, postgres, etc.)
	Type string `yaml:"type"`
	// Host is the storage host
	Host string `yaml:"host"`
	// Port is the storage port
	Port int `yaml:"port"`
	// Username is the storage username
	Username string `yaml:"username"`
	// Password is the storage password
	Password string `yaml:"password"`
	// SSL indicates whether to use SSL
	SSL bool `yaml:"ssl"`
}

// New creates a new storage configuration with default values.
func New() *Config {
	return &Config{
		Type:     DefaultType,
		Host:     DefaultHost,
		Port:     DefaultPort,
		Username: DefaultUsername,
		Password: DefaultPassword,
		SSL:      DefaultSSL,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("storage configuration is required")
	}

	if c.Type == "" {
		return errors.New("storage type is required")
	}

	if c.Host == "" {
		return errors.New("storage host is required")
	}

	if c.Port <= 0 {
		return fmt.Errorf("storage port must be positive, got %d", c.Port)
	}

	return nil
}
