// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// processorTimeout is the timeout for each processor
	processorTimeout = 30 * time.Second
	// crawlerTimeout is the timeout for waiting for the crawler to complete
	crawlerTimeout = 5 * time.Minute
	// shutdownTimeout is the timeout for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

// Cmd represents the crawl command
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Trim quotes from source name
		sourceName := strings.Trim(args[0], "\"")

		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Initialize the Fx application
		fxApp := fx.New(
			Module,
			// Override context and source name
			fx.Provide(
				func() context.Context { return ctx },
				func() string { return sourceName },
			),
		)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("failed to start application: %w", err)
		}

		// Get the handler from the Fx application
		var handler *signal.SignalHandler
		if err := fx.Populate(fxApp, &handler); err != nil {
			return fmt.Errorf("failed to get signal handler: %v", err)
		}

		// Set the fx app for coordinated shutdown
		handler.SetFXApp(fxApp)

		// Wait for completion signal
		handler.Wait()

		return nil
	},
}

// Command returns the crawl command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
