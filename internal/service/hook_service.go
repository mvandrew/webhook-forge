package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"webhook-forge/internal/domain"
	"webhook-forge/pkg/logger"
)

// HookService implements the domain.HookService interface
type HookService struct {
	repo     domain.HookRepository
	flagsDir string
	logger   logger.Logger
}

// NewHookService creates a new HookService
func NewHookService(repo domain.HookRepository, flagsDir string, logger logger.Logger) *HookService {
	return &HookService{
		repo:     repo,
		flagsDir: flagsDir,
		logger:   logger,
	}
}

// GetHook returns a hook by ID
func (s *HookService) GetHook(id string) (*domain.Hook, error) {
	hook, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}
	return hook, nil
}

// GetAllHooks returns all hooks
func (s *HookService) GetAllHooks() ([]*domain.Hook, error) {
	hooks, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("Failed to get all hooks", logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}
	return hooks, nil
}

// CreateHook creates a new hook
func (s *HookService) CreateHook(hook *domain.Hook) error {
	// Validate hook
	if err := s.validateHook(hook); err != nil {
		s.logger.Error("Failed to validate hook", logger.Field{Key: "id", Value: hook.ID}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	// Create hook
	if err := s.repo.Create(hook); err != nil {
		s.logger.Error("Failed to create hook", logger.Field{Key: "id", Value: hook.ID}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.logger.Info("Hook created", logger.Field{Key: "id", Value: hook.ID})
	return nil
}

// UpdateHook updates an existing hook
func (s *HookService) UpdateHook(hook *domain.Hook) error {
	// Validate hook
	if err := s.validateHook(hook); err != nil {
		s.logger.Error("Failed to validate hook", logger.Field{Key: "id", Value: hook.ID}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	// Update hook
	if err := s.repo.Update(hook); err != nil {
		s.logger.Error("Failed to update hook", logger.Field{Key: "id", Value: hook.ID}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.logger.Info("Hook updated", logger.Field{Key: "id", Value: hook.ID})
	return nil
}

// DeleteHook deletes a hook
func (s *HookService) DeleteHook(id string) error {
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.logger.Info("Hook deleted", logger.Field{Key: "id", Value: id})
	return nil
}

// ValidateHookToken validates a hook token
func (s *HookService) ValidateHookToken(id string, token string) error {
	hook, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get hook for token validation", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	// Check if hook is enabled
	if !hook.Enabled {
		s.logger.Warn("Hook is disabled", logger.Field{Key: "id", Value: id})
		return fmt.Errorf("hook is disabled")
	}

	// Compare tokens securely
	if subtle.ConstantTimeCompare([]byte(hook.Token), []byte(token)) != 1 {
		s.logger.Warn("Invalid token", logger.Field{Key: "id", Value: id})
		return domain.ErrInvalidToken
	}

	return nil
}

// TriggerHook triggers a hook
func (s *HookService) TriggerHook(id string, token string) error {
	// Validate token
	if err := s.ValidateHookToken(id, token); err != nil {
		return err
	}

	// Get hook
	hook, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Create flag file
	if err := s.createFlagFile(hook); err != nil {
		s.logger.Error("Failed to create flag file",
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "flag_file", Value: hook.FlagFile},
			logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.logger.Info("Hook triggered",
		logger.Field{Key: "id", Value: id},
		logger.Field{Key: "flag_file", Value: hook.FlagFile})
	return nil
}

// GenerateToken generates a random token using current time and random bytes
func (s *HookService) GenerateToken() string {
	// Get current time as part of the token generation
	timestamp := time.Now().UnixNano()

	// Create a random component (16 bytes = 32 hex chars)
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// If there's an error reading random, fallback to less random but still useful method
		s.logger.Error("Failed to generate random bytes for token", logger.Field{Key: "error", Value: err.Error()})
		randomBytes = []byte(fmt.Sprintf("%016x", timestamp))
	}

	// Combine timestamp and random component
	token := fmt.Sprintf("%x-%s", timestamp, hex.EncodeToString(randomBytes))

	return token
}

// validateHook validates a hook configuration
func (s *HookService) validateHook(hook *domain.Hook) error {
	// Check required fields
	if hook.ID == "" {
		return fmt.Errorf("hook ID is required")
	}
	if hook.Name == "" {
		return fmt.Errorf("hook name is required")
	}
	// Token validation is handled by the handler now
	if hook.FlagFile == "" {
		return fmt.Errorf("hook flag file is required")
	}

	// Validate flag file path
	if filepath.IsAbs(hook.FlagFile) {
		return fmt.Errorf("flag file path must be relative: %s", hook.FlagFile)
	}

	// Check for path traversal
	if strings.Contains(hook.FlagFile, "..") {
		return fmt.Errorf("flag file path must not contain '..': %s", hook.FlagFile)
	}

	return nil
}

// createFlagFile creates a flag file for a hook
func (s *HookService) createFlagFile(hook *domain.Hook) error {
	// Validate flag file path
	if filepath.IsAbs(hook.FlagFile) {
		return fmt.Errorf("flag file path must be relative: %s", hook.FlagFile)
	}

	// Check for path traversal
	if strings.Contains(hook.FlagFile, "..") {
		return fmt.Errorf("flag file path must not contain '..': %s", hook.FlagFile)
	}

	// Create absolute path
	flagFile := filepath.Join(s.flagsDir, hook.FlagFile)

	// Create directories
	dir := filepath.Dir(flagFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(flagFile)
	if err != nil {
		return fmt.Errorf("failed to create flag file: %w", err)
	}
	defer file.Close()

	// Write timestamp to file
	_, err = fmt.Fprintf(file, "Hook triggered at %s\n", time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to write to flag file: %w", err)
	}

	return nil
}
