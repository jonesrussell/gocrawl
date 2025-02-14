package main

import (
	"log"
	"net/http"

	"github.com/jonesrussell/gocrawl/cmd"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

func main() {
	// Initialize viper to read configuration
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Read in the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	viper.AutomaticEnv()

	// Create the configuration
	cfg, err := config.NewConfig(http.DefaultTransport)
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	// Initialize the logger
	lgr, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	// Execute the command
	if err := cmd.Execute(); err != nil {
		lgr.Fatal("Error executing command", err)
	}

	lgr.Info("Application shutdown successfully")
}
