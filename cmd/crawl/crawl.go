// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Cmd represents the crawl command.
var Cmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a website for content",
	Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
	Args: cobra.ExactArgs(1),
	RunE: runCrawl,
}

// runCrawl is the main logic of the crawl command.
func runCrawl(cmd *cobra.Command, args []string) error {
	sourceName := strings.Trim(args[0], "\"")
	cmdCtx := cmd.Context()

	// Create a new context for the crawler and Fx application.
	crawlCtx, cancel := context.WithCancel(cmdCtx)
	defer cancel()

	// Set up signal handling.
	handler := signal.NewSignalHandler(logger.NewNoOp())
	cleanup := handler.Setup(crawlCtx)
	defer cleanup()

	// Initialize the Fx application.
	fxApp := fx.New(
		Module,
		fx.Provide(
			fx.Annotate(
				func() context.Context { return crawlCtx },
				fx.ResultTags(`name:"crawlContext"`),
			),
			func() string { return sourceName },
		),
	)

	// Handle coordinated shutdown.
	handler.SetFXApp(fxApp)

	// Start the Fx application.
	if err := fxApp.Start(crawlCtx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Create a channel to signal completion
	done := make(chan error, 1)

	// Start a goroutine to wait for context cancellation
	go func() {
		select {
		case <-cmdCtx.Done():
			cancel()
			done <- fmt.Errorf("command context cancelled: %w", cmdCtx.Err())
		case <-crawlCtx.Done():
			done <- fmt.Errorf("crawl context finished: %w", crawlCtx.Err())
		}
	}()

	// Wait for either context cancellation or shutdown completion
	select {
	case err := <-done:
		return err
	default:
		return handler.Wait()
	}
}

// Command returns the crawl command for use in the root command.
func Command() *cobra.Command {
	return Cmd
}
