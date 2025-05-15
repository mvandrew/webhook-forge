package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"webhook-forge/internal/config"
	"webhook-forge/internal/service"
	"webhook-forge/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.Default()

	// Load configuration
	configPath := filepath.Join("config", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load configuration", logger.Field{Key: "error", Value: err.Error()})
	}

	// Create hook service to use for token generation
	tokenGenerator := service.NewTokenGenerator(log)

	// Generate new admin token
	newToken := tokenGenerator.GenerateToken()

	fmt.Println("Generated new admin token:")
	fmt.Println(newToken)

	// Check if token already exists in configuration
	currentToken := cfg.Server.AdminToken
	if currentToken != "" {
		fmt.Println("\nCurrent admin token:")
		fmt.Println(currentToken)
	}

	// Ask for confirmation
	fmt.Print("\nDo you want to save this new token to the configuration? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read user input", logger.Field{Key: "error", Value: err.Error()})
	}

	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("Token generation canceled. No changes made to configuration.")
		return
	}

	// Update configuration
	cfg.Server.AdminToken = newToken
	if err := cfg.Save(configPath); err != nil {
		log.Fatal("Failed to save configuration", logger.Field{Key: "error", Value: err.Error()})
	}

	fmt.Println("New admin token saved successfully to", configPath)
}
