package indices

import (
	"context"
	"errors"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var deleteSourceName string

type deleteParams struct {
	ctx     context.Context
	storage common.Storage
	sources common.Sources
	logger  common.Logger
	indices []string
	force   bool
}

// deleteCommand returns the delete command
func deleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [indices...]",
		Short: "Delete one or more Elasticsearch indices",
		Long: `Delete one or more Elasticsearch indices from the cluster.
If --source is specified, deletes the indices associated with that source.

Example:
  gocrawl indices delete my_index
  gocrawl indices delete index1 index2 index3
  gocrawl indices delete --source "Elliot Lake Today"`,
		Args: validateDeleteArgs,
		Run:  runDelete,
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&deleteSourceName, "source", "", "Delete indices for a specific source")

	return cmd
}

func validateDeleteArgs(_ *cobra.Command, args []string) error {
	if deleteSourceName == "" && len(args) == 0 {
		return errors.New("either specify indices or use --source flag")
	}
	if deleteSourceName != "" && len(args) > 0 {
		return errors.New("cannot specify both indices and --source flag")
	}
	return nil
}

func runDelete(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	var logger common.Logger

	app := fx.New(
		common.Module,
		fx.Invoke(func(storage common.Storage, sources common.Sources, l common.Logger) {
			logger = l
			params := &deleteParams{
				ctx:     context.Background(),
				storage: storage,
				sources: sources,
				logger:  l,
				indices: args,
				force:   force,
			}
			if err := executeDelete(params); err != nil {
				l.Error("Error executing delete", "error", err)
				os.Exit(1)
			}
		}),
	)

	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		if logger != nil {
			logger.Error("Error starting application", "error", err)
		}
		os.Exit(1)
	}

	if err := app.Stop(ctx); err != nil {
		if logger != nil {
			logger.Error("Error stopping application", "error", err)
		}
		os.Exit(1)
	}
}

func executeDelete(p *deleteParams) error {
	if err := resolveIndices(p); err != nil {
		return err
	}

	existingMap, err := getExistingIndices(p)
	if err != nil {
		return err
	}

	indicesToDelete := filterExistingIndices(p, existingMap)
	if len(indicesToDelete) == 0 {
		return nil
	}

	if !p.force {
		if !confirmDeletion(indicesToDelete) {
			return nil
		}
	}

	return deleteIndices(p, indicesToDelete)
}

func resolveIndices(p *deleteParams) error {
	if deleteSourceName != "" {
		source, err := p.sources.FindByName(deleteSourceName)
		if err != nil {
			return err
		}
		p.indices = []string{source.ArticleIndex, source.Index}
	}
	return nil
}

func getExistingIndices(p *deleteParams) (map[string]bool, error) {
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, idx := range indices {
		existingMap[idx] = true
	}
	return existingMap, nil
}

func filterExistingIndices(p *deleteParams, existingMap map[string]bool) []string {
	var missingIndices []string
	var indicesToDelete []string

	for _, index := range p.indices {
		if !existingMap[index] {
			missingIndices = append(missingIndices, index)
		} else {
			indicesToDelete = append(indicesToDelete, index)
		}
	}

	if len(missingIndices) > 0 {
		common.PrintInfo("\nThe following indices do not exist (already deleted):")
		for _, index := range missingIndices {
			common.PrintInfo("  - %s", index)
		}
	}

	return indicesToDelete
}

func confirmDeletion(indices []string) bool {
	common.PrintInfo("\nAre you sure you want to delete the following indices?")
	for _, index := range indices {
		common.PrintInfo("  - %s", index)
	}
	return common.PrintConfirmation("\nContinue?")
}

func deleteIndices(p *deleteParams, indices []string) error {
	for _, index := range indices {
		if err := p.storage.DeleteIndex(p.ctx, index); err != nil {
			return err
		}
		p.logger.Info("Deleted index", "index", index)
		common.PrintSuccess("Successfully deleted index '%s'", index)
	}
	return nil
}
