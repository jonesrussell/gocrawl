package logging

import (
	"errors"
	"fmt"
	"strings"
)

// Default configuration values
const (
	DefaultLevel      = "debug"
	DefaultEncoding   = "json"
	DefaultOutput     = "stdout"
	DefaultDebug      = false
	DefaultCaller     = false
	DefaultStacktrace = false
	defaultMaxSize    = 100 // 100MB
	defaultMaxBackups = 3   // Keep 3 backups
	defaultMaxAge     = 30  // Keep for 30 days
)

// Config holds logging-specific configuration settings.
type Config struct {
	// Level is the logging level (debug, info, warn, error)
	Level string `yaml:"level"`
	// Encoding is the log encoding format (json, console)
	Encoding string `yaml:"encoding"`
	// Output is the log output destination (stdout, stderr, file)
	Output string `yaml:"output"`
	// File is the log file path (only used when output is file)
	File string `yaml:"file"`
	// Debug enables debug mode for additional logging
	Debug bool `yaml:"debug"`
	// Caller enables caller information in logs
	Caller bool `yaml:"caller"`
	// Stacktrace enables stacktrace in error logs
	Stacktrace bool `yaml:"stacktrace"`
	// MaxSize is the maximum size of the log file in megabytes
	MaxSize int `yaml:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `yaml:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `yaml:"max_age"`
	// Compress determines if the rotated log files should be compressed
	Compress bool `yaml:"compress"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("log configuration is required")
	}

	if c.Level == "" {
		return errors.New("log level cannot be empty")
	}

	switch strings.ToLower(c.Level) {
	case "debug", "info", "warn", "error":
		// Valid level
	default:
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Encoding == "" {
		return errors.New("log encoding cannot be empty")
	}

	switch strings.ToLower(c.Encoding) {
	case "json", "console":
		// Valid encoding
	default:
		return fmt.Errorf("invalid log encoding: %s", c.Encoding)
	}

	if c.Output == "" {
		return errors.New("log output cannot be empty")
	}

	switch strings.ToLower(c.Output) {
	case "stdout", "stderr", "file":
		// Valid output
	default:
		return fmt.Errorf("invalid log output: %s", c.Output)
	}

	if strings.EqualFold(c.Output, "file") && c.File == "" {
		return errors.New("log file path is required when output is file")
	}

	if c.MaxSize < 0 {
		return fmt.Errorf("max size must be greater than or equal to 0, got %d", c.MaxSize)
	}

	if c.MaxBackups < 0 {
		return fmt.Errorf("max backups must be greater than or equal to 0, got %d", c.MaxBackups)
	}

	if c.MaxAge < 0 {
		return fmt.Errorf("max age must be greater than or equal to 0, got %d", c.MaxAge)
	}

	return nil
}

// New creates a new logging configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		Level:      DefaultLevel,
		Encoding:   DefaultEncoding,
		Output:     DefaultOutput,
		Debug:      DefaultDebug,
		Caller:     DefaultCaller,
		Stacktrace: DefaultStacktrace,
		MaxSize:    defaultMaxSize,
		MaxBackups: defaultMaxBackups,
		MaxAge:     defaultMaxAge,
		Compress:   true, // Compress rotated files
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures a logging configuration.
type Option func(*Config)

// WithLevel sets the log level.
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithEncoding sets the log encoding.
func WithEncoding(encoding string) Option {
	return func(c *Config) {
		c.Encoding = encoding
	}
}

// WithOutput sets the log output.
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

// WithDebug sets the debug mode.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithCaller sets whether to include caller information.
func WithCaller(caller bool) Option {
	return func(c *Config) {
		c.Caller = caller
	}
}

// WithStacktrace sets whether to include stacktrace in error logs.
func WithStacktrace(stacktrace bool) Option {
	return func(c *Config) {
		c.Stacktrace = stacktrace
	}
}

// WithMaxSize sets the maximum log file size.
func WithMaxSize(size int) Option {
	return func(c *Config) {
		c.MaxSize = size
	}
}

// WithMaxBackups sets the maximum number of log file backups.
func WithMaxBackups(backups int) Option {
	return func(c *Config) {
		c.MaxBackups = backups
	}
}

// WithMaxAge sets the maximum age of log files.
func WithMaxAge(age int) Option {
	return func(c *Config) {
		c.MaxAge = age
	}
}

// WithCompress sets whether to compress rotated log files.
func WithCompress(compress bool) Option {
	return func(c *Config) {
		c.Compress = compress
	}
}
