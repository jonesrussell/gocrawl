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
	DefaultMaxSize  = 1024 * 1024 * 1024 // 1 GB
	DefaultMaxItems = 10000
)

// Config represents storage-specific configuration settings.
type Config struct {
	// Type is the storage type (memory, file, elasticsearch)
	Type string `yaml:"type"`
	// Path is the path to the storage directory (for file storage)
	Path string `yaml:"path"`
	// MaxSize is the maximum size of the storage in bytes
	MaxSize int64 `yaml:"max_size"`
	// MaxItems is the maximum number of items to store
	MaxItems int `yaml:"max_items"`
	// Compression is whether to compress stored data
	Compression bool `yaml:"compression"`
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

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Type:        DefaultType,
		Path:        "./data",
		MaxSize:     DefaultMaxSize,
		MaxItems:    DefaultMaxItems,
		Compression: true,
		Host:        DefaultHost,
		Port:        DefaultPort,
		Username:    DefaultUsername,
		Password:    DefaultPassword,
		SSL:         DefaultSSL,
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
