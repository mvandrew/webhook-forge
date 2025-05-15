package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"webhook-forge/internal/api"
	"webhook-forge/internal/config"
	"webhook-forge/internal/middleware"
	"webhook-forge/internal/service"
	"webhook-forge/internal/storage"
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

	// Create logger with file rotation
	logConfig := logger.LogConfig{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		FilePath:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
	}

	log, err := logger.NewWithConfig(logConfig)
	if err != nil {
		fmt.Printf("Error initializing logger: %s\n", err)
		os.Exit(1)
	}
	defer log.Close()

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

	// Verify that admin token is set
	if cfg.Server.AdminToken == "" {
		log.Fatal("Admin token is not set", logger.Field{Key: "error", Value: "AdminToken is required for secure operation"})
	}

	// Create API handler
	handler := api.NewHandler(hookService, log, cfg.Server.BasePath, cfg.Server.AdminToken)

	// Create HTTP server
	mux := http.NewServeMux()

	// Create middlewares
	requestLogger := middleware.NewRequestLogger(log)
	adminAuth := middleware.NewAdminAuth(log, cfg.Server.AdminToken)
	webhookAuth := middleware.NewWebhookAuth(log, hookService)

	log.Info("Initialized authentication middlewares")

	// Set up API routes with admin authentication
	apiRoutes := handler.GetAPIRoutes()
	apiRoutesWithAuth := adminAuth.Middleware(apiRoutes)

	// Set up webhook routes with webhook authentication
	webhookRoutes := handler.GetWebhookRoutes()
	webhookRoutesWithAuth := webhookAuth.Middleware(webhookRoutes)

	// Configure routes with proper base paths
	apiPath := cfg.Server.BasePath + "/api"
	webhookPath := cfg.Server.BasePath + "/webhook"

	// Ensure paths are properly formatted
	if apiPath != "" && !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}
	if webhookPath != "" && !strings.HasPrefix(webhookPath, "/") {
		webhookPath = "/" + webhookPath
	}
	apiPath = strings.TrimSuffix(apiPath, "/")
	webhookPath = strings.TrimSuffix(webhookPath, "/")

	// Register routes with authentication middleware applied
	mux.Handle(apiPath+"/", http.StripPrefix(apiPath, apiRoutesWithAuth))
	mux.Handle(webhookPath+"/", http.StripPrefix(webhookPath, webhookRoutesWithAuth))

	// Apply request logging middleware to all requests
	middlewareChain := requestLogger.Middleware(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      middlewareChain,
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
