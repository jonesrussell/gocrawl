package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/cmd"
	"github.com/jonesrussell/gocrawl/internal/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create the configuration
	cfg, err := config.NewConfig(http.DefaultTransport)
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	// Initialize your application with the config
	cmd.Execute(cfg)
}
