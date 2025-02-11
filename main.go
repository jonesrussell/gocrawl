package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	go func() {
		if err := cmd.Execute(cfg); err != nil {
			log.Fatalf("Error executing command: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Printf("Received signal: %v. Shutting down gracefully...", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := cmd.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}

	log.Println("Application shutdown successfully")
}
