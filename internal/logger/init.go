package logger

import (
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// provideLogger sets up the global logger based on the configuration
func provideLogger(cfg config.Interface) (Interface, error) {
	appConfig := cfg.GetAppConfig()
	logConfig := cfg.GetLogConfig()

	env := appConfig.Environment
	if env == "" {
		env = "development" // Set a default environment
	}

	// Ensure we have a valid log level
	logLevel := logConfig.Level
	if logLevel == "" {
		logLevel = defaultLogLevel // Set a default log level
	}

	fmt.Printf("Initializing logger with environment: %s, log level: %s, debug: %v\n", env, logLevel, logConfig.Debug)

	var customLogger *CustomLogger
	var err error
	if env == "development" {
		customLogger, err = NewDevelopmentLogger(logLevel)
	} else {
		customLogger, err = NewProductionLogger(logLevel)
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	return customLogger, nil
}
