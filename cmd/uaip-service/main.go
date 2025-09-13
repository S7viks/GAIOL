package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gaiol/internal/models"
	"gaiol/internal/models/adapters"
)

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize model providers
	providers, err := initializeProviders(config)
	if err != nil {
		log.Fatalf("Failed to initialize providers: %v", err)
	}

	// Start the service
	if err := startService(ctx, config, providers); err != nil {
		log.Fatalf("Service error: %v", err)
	}
}

func loadConfig() (map[string]interface{}, error) {
	// Implementation for loading configuration from environment/files
	return make(map[string]interface{}), nil
}

func initializeProviders(config map[string]interface{}) (map[string]models.ModelProvider, error) {
	providers := make(map[string]models.ModelProvider)

	// Initialize Gemini provider
	gemini := adapters.NewGeminiProvider()
	if err := gemini.Initialize(config); err != nil {
		return nil, err
	}
	providers["gemini"] = gemini

	// Add more providers here

	return providers, nil
}

func startService(ctx context.Context, config map[string]interface{}, providers map[string]models.ModelProvider) error {
	// Implementation for starting the service
	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	return nil
}
