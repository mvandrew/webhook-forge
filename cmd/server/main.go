package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"webhook-forge/internal/api"
	"webhook-forge/internal/config"
	"webhook-forge/internal/service"
	"webhook-forge/internal/storage"
	"webhook-forge/pkg/logger"
)

func main() {
	// Load configuration
	configPath := filepath.Join("config", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	// Create logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format, os.Stdout)
	log.Info("Starting webhook-forge server")

	// Create hooks directory
	hooksDir := filepath.Dir(cfg.Hooks.StoragePath)
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		log.Fatal("Failed to create hooks directory", logger.Field{Key: "directory", Value: hooksDir}, logger.Field{Key: "error", Value: err.Error()})
	}

	// Create flags directory
	if err := os.MkdirAll(cfg.Hooks.FlagsDir, 0755); err != nil {
		log.Fatal("Failed to create flags directory", logger.Field{Key: "directory", Value: cfg.Hooks.FlagsDir}, logger.Field{Key: "error", Value: err.Error()})
	}

	// Create hook repository
	hookRepo, err := storage.NewJSONHookRepository(cfg.Hooks.StoragePath)
	if err != nil {
		log.Fatal("Failed to create hook repository", logger.Field{Key: "error", Value: err.Error()})
	}

	// Create hook service
	hookService := service.NewHookService(hookRepo, cfg.Hooks.FlagsDir, log)

	// Create API handler
	handler := api.NewHandler(hookService, log, cfg.Server.BasePath)

	// Create HTTP server
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTP server", logger.Field{Key: "address", Value: addr}, logger.Field{Key: "base_path", Value: cfg.Server.BasePath})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", logger.Field{Key: "error", Value: err.Error()})
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error", logger.Field{Key: "error", Value: err.Error()})
	}

	log.Info("Server stopped gracefully")
}
