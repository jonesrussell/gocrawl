// Package config provides configuration management for the application.
package config

import "errors"

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
