package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/cmd"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

func main() {
	// Initialize viper to read configuration
	viper.SetConfigName(".env") // Name of the env file (without extension)
	viper.SetConfigType("env")  // The type of the file, here it is env
	viper.AddConfigPath(".")    // Path to look for the env file in the current directory

	// Read in the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	viper.AutomaticEnv() // Load environment variables

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

	// Initialize your application with the config and logger
	go func() {
		if err := cmd.Execute(); err != nil {
			lgr.Fatal("Error executing command", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	lgr.Info("Received signal: %v. Shutting down gracefully...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := cmd.Shutdown(ctx); err != nil {
		lgr.Fatal("Error during shutdown", err)
	}

	lgr.Info("Application shutdown successfully")
}
