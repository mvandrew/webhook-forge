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
	// Load configuration
	// Check if CONFIG_PATH environment variable is set
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = filepath.Join("config", "config.json")
	}
	fmt.Printf("Loading configuration from: %s\n", configPath)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	// Initialize logger with file rotation
	var log logger.Logger

	if cfg != nil && cfg.Log.FilePath != "" {
		logConfig := logger.LogConfig{
			Level:      cfg.Log.Level,
			Format:     cfg.Log.Format,
			FilePath:   cfg.Log.FilePath,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
		}

		log, err = logger.NewWithConfig(logConfig)
		if err != nil {
			fmt.Printf("Error initializing logger: %s\n", err)
			log = logger.Default() // Fallback to default logger
		}
	} else {
		log = logger.Default()
	}
	defer log.Close()

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
