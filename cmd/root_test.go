package cmd

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// MockConfig is a mock implementation of the config.Config interface
type MockConfig struct {
	App     config.AppConfig
	Crawler struct {
		BaseURL   string
		IndexName string
		MaxDepth  int
	}
}

func NewMockConfig() *config.Config {
	// Create and return a mock config that matches the expected structure
	return &config.Config{
		App: config.AppConfig{
			Environment: "development",
		},
		Crawler: config.CrawlerConfig{
			BaseURL:   "http://example.com",
			IndexName: "example_index",
			MaxDepth:  3,
		},
	}
}

func TestCommandsRegistered(t *testing.T) {
	// Set the config path to the directory where your config file is located
	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // Set the type of the config file
	viper.AddConfigPath("../..")  // Adjust the path to where the config file is located
	viper.AutomaticEnv()          // Automatically read environment variables

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the logger and config for testing
	log, err := logger.NewDevelopmentLogger() // or use a mock logger
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	cfg := NewMockConfig() // Use the mock config

	// Initialize the storage
	elasticClient, err := storage.ProvideElasticsearchClient(cfg, log) // Create the Elasticsearch client
	if err != nil {
		t.Fatalf("Failed to create Elasticsearch client: %v", err)
	}
	storageInstance, err := storage.NewStorage(elasticClient, log) // Create the storage instance
	if err != nil {
		t.Fatalf("Failed to create storage instance: %v", err)
	}

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
	crawlerInstance, err := crawler.NewCrawler(crawlerParams) // Ensure the crawler is initialized
	if err != nil {
		t.Fatalf("Failed to create crawler instance: %v", err)
	}

	// Initialize multisource
	multiSourceInstance, err := multisource.NewMultiSource(log, crawlerInstance.Crawler) // Pass logger and crawler
	if err != nil {
		t.Fatalf("Failed to create multi source instance: %v", err)
	}
	multiCrawlCmd := NewMultiCrawlCmd(log, cfg, multiSourceInstance) // Pass the instance to the command

	// Register commands
	rootCmd.AddCommand(crawlCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(multiCrawlCmd)

	// Check if commands are registered
	crawlCommand, _, err := rootCmd.Find("crawl") // Correctly use Find with a single string
	assert.NoError(t, err)
	assert.NotNil(t, crawlCommand, "Crawl command should be registered")

	searchCommand, _, err := rootCmd.Find("search") // Correctly use Find with a single string
	assert.NoError(t, err)
	assert.NotNil(t, searchCommand, "Search command should be registered")

	multiCommand, _, err := rootCmd.Find("multi") // Correctly use Find with a single string
	assert.NoError(t, err)
	assert.NotNil(t, multiCommand, "Multi command should be registered")
}
