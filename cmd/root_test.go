package cmd

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCommandsRegistered(t *testing.T) {
	// Initialize the logger and config for testing
	log, _ := logger.NewDevelopmentLogger() // or use a mock logger
	cfg, _ := config.NewConfig()            // or use a mock config

	// Initialize the storage
	elasticClient, _ := storage.ProvideElasticsearchClient(cfg, log) // Create the Elasticsearch client
	storageInstance, _ := storage.NewStorage(elasticClient, log)     // Create the storage instance

	// Initialize the commands
	crawlCmd := NewCrawlCmd(log, cfg)   // Pass only logger and config
	searchCmd := NewSearchCmd(log, cfg) // Pass logger and config

	// Initialize the debugger
	debuggerInstance := logger.NewCollyDebugger(log) // Pass the logger to the debugger

	// Initialize the crawler
	crawlerParams := crawler.Params{
		Logger:   log,
		Storage:  storageInstance,
		Debugger: debuggerInstance,
		Config:   cfg,
	}
	crawlerInstance, _ := crawler.NewCrawler(crawlerParams) // Ensure the crawler is initialized
	// Initialize multisource
	multiSourceInstance, _ := multisource.NewMultiSource(log, crawlerInstance.Crawler) // Pass logger and crawler
	multiCrawlCmd := NewMultiCrawlCmd(log, cfg, multiSourceInstance)                   // Pass the instance to the command

	// Register commands
	rootCmd.AddCommand(crawlCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(multiCrawlCmd)

	// Check if commands are registered
	crawlCommand, _, err := rootCmd.Find([]string{"crawl"})
	assert.NoError(t, err)
	assert.NotNil(t, crawlCommand, "Crawl command should be registered")

	searchCommand, _, err := rootCmd.Find([]string{"search"})
	assert.NoError(t, err)
	assert.NotNil(t, searchCommand, "Search command should be registered")

	multiCommand, _, err := rootCmd.Find([]string{"multi"})
	assert.NoError(t, err)
	assert.NotNil(t, multiCommand, "Multi command should be registered")
}
