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
	"github.com/stretchr/testify/mock"
)

// MockConfig is a mock implementation of the config.Config interface
type MockConfig struct {
	App     config.AppConfig
	Crawler config.CrawlerConfig
}

func NewMockConfig() *config.Config {
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
	viper.SetConfigName("sources") // Name of the config file (without extension)
	viper.SetConfigType("yaml")    // Set the type of the config file
	viper.AddConfigPath("..")      // Adjust the path to where the config file is located
	viper.AutomaticEnv()           // Automatically read environment variables

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the logger and config for testing
	log := logger.NewMockLogger() // or use a mock logger

	// Set up expectations for the mock logger
	log.On("Info", mock.Anything, mock.Anything).Return() // Expect Info method to be called

	cfg := NewMockConfig() // Use the mock config

	// Initialize the storage
	storageInstance := storage.NewMockStorage() // Create the storage instance

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
	multiSourceInstance, err := multisource.NewMultiSource(log, crawlerInstance.Crawler, "../sources.yml", "example") // Pass logger, crawler, and config path
	if err != nil {
		t.Fatalf("Failed to create multi source instance: %v", err)
	}
	multiCrawlCmd := NewMultiCrawlCmd(log, cfg, multiSourceInstance, crawlerInstance.Crawler) // Pass the instance to the command

	// Register commands
	rootCmd.AddCommand(crawlCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(multiCrawlCmd)

	// Check if commands are registered
	crawlCommand, _, err := rootCmd.Find([]string{"crawl"}) // Use a slice of strings
	assert.NoError(t, err)
	assert.NotNil(t, crawlCommand, "Crawl command should be registered")

	searchCommand, _, err := rootCmd.Find([]string{"search"}) // Use a slice of strings
	assert.NoError(t, err)
	assert.NotNil(t, searchCommand, "Search command should be registered")

	multiCommand, _, err := rootCmd.Find([]string{"multi"}) // Use a slice of strings
	assert.NoError(t, err)
	assert.NotNil(t, multiCommand, "Multi command should be registered")
}

func TestNewMultiCrawlCmd(t *testing.T) {
	log := logger.NewMockLogger()             // Assuming you have a mock logger
	cfg := &config.Config{}                   // Initialize your config as needed
	multiSource := &multisource.MultiSource{} // Initialize your multisource as needed
	crawlerInstance := &crawler.Crawler{}     // Initialize your crawler as needed

	// Update this line to include the crawler instance
	multiCmd := NewMultiCrawlCmd(log, cfg, multiSource, crawlerInstance)

	assert.NotNil(t, multiCmd)
}
