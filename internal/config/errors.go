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

// ValidationError represents an error in configuration validation
type ValidationError struct {
	Field  string
	Value  any
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid config: field %q with value %v: %s", e.Field, e.Value, e.Reason)
}

// LoadError represents an error loading configuration
type LoadError struct {
	File string
	Err  error
}

func (e *LoadError) Error() string {
	return fmt.Sprintf("failed to load config from %s: %v", e.File, e.Err)
}

func (e *LoadError) Unwrap() error {
	return e.Err
}

// ParseError represents an error parsing configuration
type ParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse config field %q with value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// SourceError represents an error loading source configuration
type SourceError struct {
	Source string
	Err    error
}

func (e *SourceError) Error() string {
	return fmt.Sprintf("failed to load source %s: %v", e.Source, e.Err)
}

func (e *SourceError) Unwrap() error {
	return e.Err
}

// ViperError represents an error from Viper configuration
type ViperError struct {
	Operation string
	Err       error
}

func (e *ViperError) Error() string {
	return fmt.Sprintf("viper error during %s: %v", e.Operation, e.Err)
}

func (e *ViperError) Unwrap() error {
	return e.Err
}
