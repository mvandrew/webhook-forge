package domain

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrHookNotFound      = errors.New("hook not found")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidHookConfig = errors.New("invalid hook configuration")
)

// Hook represents a webhook configuration
type Hook struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Token       string    `json:"token"`
	FlagFile    string    `json:"flag_file"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HookRepository defines the interface for hook storage
type HookRepository interface {
	GetByID(id string) (*Hook, error)
	GetAll() ([]*Hook, error)
	Create(hook *Hook) error
	Update(hook *Hook) error
	Delete(id string) error
}

// HookService defines the interface for hook business logic
type HookService interface {
	GetHook(id string) (*Hook, error)
	GetAllHooks() ([]*Hook, error)
	CreateHook(hook *Hook) error
	UpdateHook(hook *Hook) error
	DeleteHook(id string) error
	ValidateHookToken(id string, token string) error
	TriggerHook(id string, token string) error
	GenerateToken() string
}
