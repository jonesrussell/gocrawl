package cmdutil

import (
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	// cfgFile holds the path to the configuration file.
	cfgFile string
)

// NewCommand creates a new cobra command with consistent setup and configuration.
// It provides a standard way to create commands with proper documentation and config handling.
func NewCommand(use, short, long string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               use,
		Short:             short,
		Long:              long,
		PersistentPreRunE: setupConfig,
	}

	// Add the persistent --config flag to all commands
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")

	return cmd
}

// SetupFxApp creates a new fx application with common modules and configuration.
// It provides a standard way to set up dependency injection for commands.
func SetupFxApp(cmd *cobra.Command, modules ...fx.Option) *fx.App {
	return fx.New(
		append([]fx.Option{
			common.Module,
			fx.Provide(func() *cobra.Command { return cmd }),
		}, modules...)...,
	)
}

// setupConfig handles configuration file setup for all commands.
// It ensures the config file path is absolute and sets it in the environment.
func setupConfig(_ *cobra.Command, _ []string) error {
	// If config file is provided via flag, use absolute path
	if cfgFile != "" {
		if !os.IsPathSeparator(cfgFile[0]) {
			// Convert relative path to absolute
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			cfgFile = wd + string(os.PathSeparator) + cfgFile
		}
		os.Setenv("CONFIG_FILE", cfgFile)
	}
	return nil
}

// ExecuteCommand runs a command and handles any errors that occur during execution.
// It provides consistent error handling across all commands.
func ExecuteCommand(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		common.PrintErrorf("Error executing command: %v", err)
		os.Exit(1)
	}
}
