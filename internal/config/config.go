package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `json:"server"`
	Hooks  HooksConfig  `json:"hooks"`
	Log    LogConfig    `json:"log"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	BasePath string `json:"base_path"` // Base path for all routes, e.g. "/hooks" when proxied behind nginx
}

// HooksConfig contains webhook configuration
type HooksConfig struct {
	StoragePath string `json:"storage_path"`
	FlagsDir    string `json:"flags_dir"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	// Default configuration
	cfg := &Config{
		Server: ServerConfig{
			Host:     "127.0.0.1",
			Port:     8080,
			BasePath: "", // Empty string means no base path (server at root)
		},
		Hooks: HooksConfig{
			StoragePath: "data/hooks.json",
			FlagsDir:    "data/flags",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directories
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		// Save default config
		if err := cfg.Save(path); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}

		return cfg, nil
	}

	// Read config file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Decode JSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	// Create directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
