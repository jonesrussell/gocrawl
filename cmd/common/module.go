package common

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides common dependencies for all commands
var Module = fx.Module("common")

// CommandDeps holds common dependencies for commands
type CommandDeps struct {
	fx.In

	Config  config.Interface
	Logger  logger.Interface
	Context context.Context
}

// CommandRegistrar is a function type that registers a command with its parent
type CommandRegistrar func(parent *cobra.Command)
