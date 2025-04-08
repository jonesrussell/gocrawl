// Package log provides logging-related configuration types and functions.
package log

import (
	"errors"
	"fmt"
)

// Default configuration values
const (
	DefaultLevel      = "info"
	DefaultFormat     = "json"
	DefaultOutput     = "stdout"
	DefaultMaxSize    = 100
	DefaultMaxBackups = 3
	DefaultMaxAge     = 28
	DefaultCompress   = true
)

// Config holds logging-specific configuration settings.
type Config struct {
	// Level is the logging level (debug, info, warn, error)
	Level string `yaml:"level"`
	// Format is the log format (json, text)
	Format string `yaml:"format"`
	// Output is the log output destination (stdout, stderr, file)
	Output string `yaml:"output"`
	// File is the log file path (only used when output is file)
	File string `yaml:"file"`
	// MaxSize is the maximum size of the log file in megabytes
	MaxSize int `yaml:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `yaml:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `yaml:"max_age"`
	// Compress determines if the rotated log files should be compressed
	Compress bool `yaml:"compress"`
}

// New creates a new logging configuration with default values.
func New() *Config {
	return &Config{
		Level:      DefaultLevel,
		Format:     DefaultFormat,
		Output:     DefaultOutput,
		MaxSize:    DefaultMaxSize,
		MaxBackups: DefaultMaxBackups,
		MaxAge:     DefaultMaxAge,
		Compress:   DefaultCompress,
	}
}

// Option is a function that configures a logging configuration.
type Option func(*Config)

// WithLevel sets the logging level.
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithFormat sets the log format.
func WithFormat(format string) Option {
	return func(c *Config) {
		c.Format = format
	}
}

// WithOutput sets the log output destination.
func WithOutput(output string) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithFile sets the log file path.
func WithFile(file string) Option {
	return func(c *Config) {
		c.File = file
	}
}

// WithMaxSize sets the maximum log file size.
func WithMaxSize(size int) Option {
	return func(c *Config) {
		c.MaxSize = size
	}
}

// WithMaxBackups sets the maximum number of old log files to retain.
func WithMaxBackups(backups int) Option {
	return func(c *Config) {
		c.MaxBackups = backups
	}
}

// WithMaxAge sets the maximum age of old log files.
func WithMaxAge(age int) Option {
	return func(c *Config) {
		c.MaxAge = age
	}
}

// WithCompress sets whether to compress old log files.
func WithCompress(compress bool) Option {
	return func(c *Config) {
		c.Compress = compress
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("log configuration is required")
	}

	if c.Level == "" {
		return errors.New("log level is required")
	}

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Format == "" {
		return errors.New("log format is required")
	}

	if c.Output == "" {
		return errors.New("log output is required")
	}

	if c.Output == "file" && c.File == "" {
		return errors.New("log file is required when output is file")
	}

	if c.MaxSize < 0 {
		return fmt.Errorf("max size must be non-negative, got %d", c.MaxSize)
	}

	if c.MaxBackups < 0 {
		return fmt.Errorf("max backups must be non-negative, got %d", c.MaxBackups)
	}

	if c.MaxAge < 0 {
		return fmt.Errorf("max age must be non-negative, got %d", c.MaxAge)
	}

	return nil
}
