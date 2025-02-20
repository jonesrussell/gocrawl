package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var sourceName string

// createMultiCmd creates the multi command
func createMultiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		RunE:  runMultiCmd,
	}
}

// runMultiCmd is the function to execute the multi command
func runMultiCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	globalLogger.Debug("Starting multi-crawl command...", "sourceName", sourceName)

	// Create an Fx application
	app := fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		crawler.Module,
		multisource.Module,
		fx.Provide(func(c *crawler.Crawler) *multisource.MultiSource {
			ms, err := multisource.NewMultiSource(globalLogger, c, "sources.yml")
			if err != nil {
				return nil
			}
			return ms
		}),
		fx.Invoke(startMultiSourceCrawl),
	)

	// Start the application
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	defer func() {
		if err := app.Stop(ctx); err != nil {
			globalLogger.Error("Error stopping application", "context", ctx, "error", err)
		}
	}()

	return nil
}

// startMultiSourceCrawl starts the multi-source crawl
func startMultiSourceCrawl(ms *multisource.MultiSource, c *crawler.Crawler) error {
	if c == nil {
		return errors.New("crawler is not initialized")
	}

	// Filter sources based on sourceName
	filteredSources, err := filterSources(ms.Sources, sourceName)
	if err != nil {
		return err
	}

	// Extract base URL, max_depth, rate_limit, and index from the filtered source
	baseURL := filteredSources[0].URL
	maxDepth := filteredSources[0].MaxDepth
	rateLimit, err := time.ParseDuration(filteredSources[0].RateLimit) // Parse rate limit
	if err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}
	indexName := filteredSources[0].Index // Extract index name

	// Set the index name in the Crawler's configuration
	c.IndexName = indexName // Set the IndexName from the source

	// Create the collector using the collector module
	collectorResult, err := collector.New(collector.Params{
		BaseURL:   baseURL,
		MaxDepth:  maxDepth,  // Use the extracted max_depth
		RateLimit: rateLimit, // Use the extracted rate_limit
		Debugger:  logger.NewCollyDebugger(globalLogger),
		Logger:    globalLogger,
	}, c)
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	c.SetCollector(collectorResult.Collector)

	// Start the multi-source crawl
	return ms.Start(context.Background(), sourceName)
}

func init() {
	multiCmd := createMultiCmd()
	rootCmd.AddCommand(multiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// multiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// multiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	multiCmd.Flags().StringVar(&sourceName, "source", "", "Specify the source to crawl")
	if err := multiCmd.MarkFlagRequired("source"); err != nil {
		globalLogger.Error("Error marking source flag as required", "error", err)
	}
}

// filterSources filters the sources based on source name
func filterSources(sources []multisource.SourceConfig, sourceName string) ([]multisource.SourceConfig, error) {
	var filteredSources []multisource.SourceConfig
	for _, source := range sources {
		if source.Name == sourceName {
			filteredSources = append(filteredSources, source)
		}
	}
	if len(filteredSources) == 0 {
		return nil, fmt.Errorf("no source found with name: %s", sourceName)
	}
	return filteredSources, nil
}
