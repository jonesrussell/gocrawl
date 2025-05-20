// Package index provides commands for managing Elasticsearch index.
package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	forceDelete bool
)

// Command is the index command
var Command = &cobra.Command{
	Use:   "index",
	Short: "Manage Elasticsearch indices",
	Long:  `Manage Elasticsearch indices for storing crawled content`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// IndexParams contains dependencies for index commands
type IndexParams struct {
	fx.In

	Config       config.Interface
	Logger       logger.Interface
	Storage      types.Interface
	IndexManager types.IndexManager
	Sources      sources.Interface
}

// Module provides the index module for dependency injection.
var Module = fx.Module("index",
	common.Module,
	storage.Module,
	sources.Module,
	fx.Provide(
		NewTableRenderer,
		NewLister,

		// Provide the creator
		func(p IndexParams) *Creator {
			return NewCreator(p.Config, p.Logger, p.Storage, CreateParams{})
		},

		// Provide the deleter
		func(p IndexParams) *Deleter {
			return NewDeleter(p.Config, p.Logger, p.Storage, p.Sources, DeleteParams{
				Force: forceDelete,
			})
		},
	),
)

// init initializes the index command and its subcommands
func init() {
	Command.AddCommand(createListCmd(), createCreateCmd(), createDeleteCmd())
}

// createListCmd creates the list command
func createListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all indices",
		RunE:  runListCmd,
	}
}

// createCreateCmd creates the create command
func createCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [index-name]",
		Short: "Create an index",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreateCmd,
	}
}

// createDeleteCmd creates the delete command
func createDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [index-name]",
		Short: "Delete an index",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runDeleteCmd,
	}
	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion without confirmation")
	return cmd
}

// runListCmd executes the list command
func runListCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	app := fx.New(
		Module,
		fx.Provide(
			func() config.Interface { return cfg },
			func() logger.Interface { return log },
		),
		fx.Invoke(func(lister *Lister) error {
			return lister.Start(ctx)
		}),
	)

	return runApp(ctx, app, log)
}

// runCreateCmd executes the create command
func runCreateCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	app := fx.New(
		Module,
		fx.Provide(
			func() config.Interface { return cfg },
			func() logger.Interface { return log },
		),
		fx.Invoke(func(creator *Creator) error {
			creator.index = args[0]
			return creator.Start(ctx)
		}),
	)

	return runApp(ctx, app, log)
}

// runDeleteCmd executes the delete command
func runDeleteCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	app := fx.New(
		Module,
		fx.Provide(
			func() config.Interface { return cfg },
			func() logger.Interface { return log },
		),
		fx.Invoke(func(deleter *Deleter) error {
			deleter.index = args
			return deleter.Start(ctx)
		}),
	)

	return runApp(ctx, app, log)
}

// runApp starts and stops the fx application
func runApp(ctx context.Context, app *fx.App, log logger.Interface) error {
	if startErr := app.Start(ctx); startErr != nil {
		return fmt.Errorf("failed to start application: %w", startErr)
	}
	defer func() {
		if stopErr := app.Stop(ctx); stopErr != nil {
			log.Error("failed to stop application", "error", stopErr)
		}
	}()
	return nil
}

// getDependencies retrieves logger and config from context
func getDependencies(ctx context.Context) (log logger.Interface, cfg config.Interface, err error) {
	// Get logger from context
	loggerValue := ctx.Value(common.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return nil, nil, errors.New("logger not found in context")
	}

	// Get config from context
	configValue := ctx.Value(common.ConfigKey)
	cfg, ok = configValue.(config.Interface)
	if !ok {
		return nil, nil, errors.New("config not found in context")
	}

	return log, cfg, nil
}
