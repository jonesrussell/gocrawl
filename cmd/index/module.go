// Package index provides commands for managing Elasticsearch index.
package index

import (
	"context"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
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
	sourceName  string
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
		Use:   "delete [index-name...]",
		Short: "Delete an index",
		Long:  `Delete one or more indices. Either provide index names as arguments or use --source to delete indices for a specific source.`,
		Args:  cobra.MinimumNArgs(0),
		RunE:  runDeleteCmd,
	}
	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion without confirmation")
	cmd.Flags().StringVar(&sourceName, "source", "", "Delete index for a specific source by name")
	return cmd
}

// runListCmd executes the list command
func runListCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	// Construct dependencies directly without FX
	storageClient, err := createStorageClient(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: storageClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	renderer := NewTableRenderer(log)
	lister := NewLister(cfg, log, storageResult.Storage, renderer)

	return lister.Start(ctx)
}

// runCreateCmd executes the create command
func runCreateCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	// Construct dependencies directly without FX
	storageClient, err := createStorageClient(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: storageClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	creator := NewCreator(cfg, log, storageResult.Storage, CreateParams{})
	creator.index = args[0]

	return creator.Start(ctx)
}

// runDeleteCmd executes the delete command
func runDeleteCmd(cmd *cobra.Command, args []string) error {
	// Validate that either --source is provided or at least one index name is provided
	if sourceName == "" && len(args) == 0 {
		return fmt.Errorf("either --source flag or at least one index name must be provided")
	}

	ctx := cmd.Context()
	log, cfg, err := getDependencies(ctx)
	if err != nil {
		return err
	}

	// Construct dependencies directly without FX
	storageClient, err := createStorageClient(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}

	storageResult, err := storage.NewStorage(storage.StorageParams{
		Config: cfg,
		Logger: log,
		Client: storageClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	sourcesManager, err := sources.LoadSources(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	deleter := NewDeleter(cfg, log, storageResult.Storage, sourcesManager, DeleteParams{
		Force:      forceDelete,
		SourceName: sourceName,
		Indices:    args,
	})

	return deleter.Start(ctx)
}

// getDependencies retrieves logger and config from context using helper
func getDependencies(ctx context.Context) (log logger.Interface, cfg config.Interface, err error) {
	return common.GetDependencies(ctx)
}

// createStorageClient creates an Elasticsearch client without using FX
func createStorageClient(cfg config.Interface, log logger.Interface) (*es.Client, error) {
	clientResult, err := storage.NewClient(storage.ClientParams{
		Config: cfg,
		Logger: log,
	})
	if err != nil {
		return nil, err
	}
	return clientResult.Client, nil
}
