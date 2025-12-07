package common

import (
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

// NewCommandDeps creates CommandDeps by loading config and creating logger.
// This consolidates the common initialization code from Execute().
func NewCommandDeps() (CommandDeps, error) {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return CommandDeps{}, fmt.Errorf("load config: %w", err)
	}

	// Get logger configuration from Viper
	logLevel := viper.GetString("logger.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logLevel = strings.ToLower(logLevel)

	logCfg := &logger.Config{
		Level:       logger.Level(logLevel),
		Development: viper.GetBool("logger.development"),
		Encoding:    viper.GetString("logger.encoding"),
		OutputPaths: viper.GetStringSlice("logger.output_paths"),
		EnableColor: viper.GetBool("logger.enable_color"),
	}

	log, err := logger.New(logCfg)
	if err != nil {
		return CommandDeps{}, fmt.Errorf("create logger: %w", err)
	}

	deps := CommandDeps{
		Logger: log,
		Config: cfg,
	}

	if validateErr := deps.Validate(); validateErr != nil {
		return CommandDeps{}, fmt.Errorf("validate deps: %w", validateErr)
	}

	return deps, nil
}
