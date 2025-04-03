// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
//
// The package is organized into several files:
//   - interface.go: Defines the Interface for configuration management
//   - types.go: Contains all configuration-related types
//   - constants.go: Contains default values and constants
//   - errors.go: Contains error definitions
//   - config.go: Contains the main configuration implementation
//   - validate.go: Contains validation logic
//   - selectors.go: Contains selector-related types and functions
//   - module.go: Contains dependency injection setup
//
// Example usage:
//
//	cfg, err := config.LoadConfig("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access configuration values
//	appConfig := cfg.GetAppConfig()
//	logConfig := cfg.GetLogConfig()
//	crawlerConfig := cfg.GetCrawlerConfig()
//
// Configuration can be loaded from:
//   - YAML files
//   - Environment variables
//   - Command line flags
//
// The configuration is validated on load to ensure all required values
// are present and have valid types.
package config
