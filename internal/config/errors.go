// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
)

// Common configuration errors
var (
	// ErrConfigNotFound is returned when the configuration file is not found
	ErrConfigNotFound = errors.New("configuration file not found")

	// ErrConfigInvalid is returned when the configuration is invalid
	ErrConfigInvalid = errors.New("invalid configuration")

	// ErrConfigLoadFailed is returned when loading the configuration fails
	ErrConfigLoadFailed = errors.New("failed to load configuration")

	// ErrConfigSaveFailed is returned when saving the configuration fails
	ErrConfigSaveFailed = errors.New("failed to save configuration")

	// ErrConfigValidationFailed is returned when configuration validation fails
	ErrConfigValidationFailed = errors.New("configuration validation failed")

	// ErrConfigParseFailed is returned when parsing the configuration fails
	ErrConfigParseFailed = errors.New("failed to parse configuration")

	// ErrConfigEnvVarNotFound is returned when a required environment variable is not found
	ErrConfigEnvVarNotFound = errors.New("required environment variable not found")

	// ErrConfigEnvVarInvalid is returned when an environment variable has an invalid value
	ErrConfigEnvVarInvalid = errors.New("invalid environment variable value")

	// ErrConfigFileNotFound is returned when a required configuration file is not found
	ErrConfigFileNotFound = errors.New("required configuration file not found")

	// ErrConfigFileInvalid is returned when a configuration file has invalid content
	ErrConfigFileInvalid = errors.New("invalid configuration file content")
)

// ConfigValidationError represents an error in configuration validation
type ConfigValidationError struct {
	Field  string
	Value  any
	Reason string
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("invalid config: field %q with value %v: %s", e.Field, e.Value, e.Reason)
}

// ConfigLoadError represents an error loading configuration
type ConfigLoadError struct {
	File string
	Err  error
}

func (e *ConfigLoadError) Error() string {
	return fmt.Sprintf("failed to load config from %s: %v", e.File, e.Err)
}

func (e *ConfigLoadError) Unwrap() error {
	return e.Err
}

// ConfigParseError represents an error parsing configuration
type ConfigParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ConfigParseError) Error() string {
	return fmt.Sprintf("failed to parse config field %q with value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ConfigParseError) Unwrap() error {
	return e.Err
}

// ConfigSourceError represents an error loading source configuration
type ConfigSourceError struct {
	Source string
	Err    error
}

func (e *ConfigSourceError) Error() string {
	return fmt.Sprintf("failed to load source %s: %v", e.Source, e.Err)
}

func (e *ConfigSourceError) Unwrap() error {
	return e.Err
}

// ConfigViperError represents an error from Viper configuration
type ConfigViperError struct {
	Operation string
	Err       error
}

func (e *ConfigViperError) Error() string {
	return fmt.Sprintf("viper error during %s: %v", e.Operation, e.Err)
}

func (e *ConfigViperError) Unwrap() error {
	return e.Err
}
