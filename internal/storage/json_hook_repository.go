package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"webhook-forge/internal/domain"
)

// JSONHookRepository implements the HookRepository interface with JSON file storage
type JSONHookRepository struct {
	filePath string
	hooks    map[string]*domain.Hook
	mu       sync.RWMutex
}

// NewJSONHookRepository creates a new JSONHookRepository
func NewJSONHookRepository(filePath string) (*JSONHookRepository, error) {
	repo := &JSONHookRepository{
		filePath: filePath,
		hooks:    make(map[string]*domain.Hook),
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create empty file
		if err := repo.save(); err != nil {
			return nil, fmt.Errorf("failed to create hooks file: %w", err)
		}
	} else {
		// Load existing hooks
		if err := repo.load(); err != nil {
			return nil, fmt.Errorf("failed to load hooks: %w", err)
		}
	}

	return repo, nil
}

// GetByID returns a hook by ID
func (r *JSONHookRepository) GetByID(id string) (*domain.Hook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hook, ok := r.hooks[id]
	if !ok {
		return nil, domain.ErrHookNotFound
	}

	return hook, nil
}

// GetAll returns all hooks
func (r *JSONHookRepository) GetAll() ([]*domain.Hook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hooks := make([]*domain.Hook, 0, len(r.hooks))
	for _, hook := range r.hooks {
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

// Create creates a new hook
func (r *JSONHookRepository) Create(hook *domain.Hook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if hook already exists
	if _, ok := r.hooks[hook.ID]; ok {
		return fmt.Errorf("hook with ID %s already exists", hook.ID)
	}

	// Set creation time
	now := time.Now()
	hook.CreatedAt = now
	hook.UpdatedAt = now

	// Add hook
	r.hooks[hook.ID] = hook

	// Save hooks
	return r.save()
}

// Update updates an existing hook
func (r *JSONHookRepository) Update(hook *domain.Hook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if hook exists
	if _, ok := r.hooks[hook.ID]; !ok {
		return domain.ErrHookNotFound
	}

	// Update time
	hook.UpdatedAt = time.Now()

	// Update hook
	r.hooks[hook.ID] = hook

	// Save hooks
	return r.save()
}

// Delete deletes a hook
func (r *JSONHookRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if hook exists
	if _, ok := r.hooks[id]; !ok {
		return domain.ErrHookNotFound
	}

	// Delete hook
	delete(r.hooks, id)

	// Save hooks
	return r.save()
}

// load loads hooks from file
func (r *JSONHookRepository) load() error {
	// Open file
	file, err := os.Open(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to open hooks file: %w", err)
	}
	defer file.Close()

	// Decode JSON
	var hooks []*domain.Hook
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&hooks); err != nil {
		// If file is empty, return empty hooks
		if err.Error() == "EOF" {
			return nil
		}
		return fmt.Errorf("failed to decode hooks file: %w", err)
	}

	// Add hooks to map
	for _, hook := range hooks {
		r.hooks[hook.ID] = hook
	}

	return nil
}

// save saves hooks to file
func (r *JSONHookRepository) save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file
	file, err := os.Create(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to create hooks file: %w", err)
	}
	defer file.Close()

	// Convert map to slice
	hooks := make([]*domain.Hook, 0, len(r.hooks))
	for _, hook := range r.hooks {
		hooks = append(hooks, hook)
	}

	// Encode JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(hooks); err != nil {
		return fmt.Errorf("failed to encode hooks: %w", err)
	}

	return nil
}
