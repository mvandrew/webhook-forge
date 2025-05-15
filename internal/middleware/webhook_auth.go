package middleware

import (
	"net/http"
	"strings"

	"webhook-forge/internal/domain"
	"webhook-forge/pkg/logger"
)

// WebhookAuth provides middleware for webhook authentication
type WebhookAuth struct {
	logger      logger.Logger
	hookService domain.HookService
}

// NewWebhookAuth creates a new webhook authentication middleware
func NewWebhookAuth(logger logger.Logger, hookService domain.HookService) domain.WebhookAuthMiddleware {
	return &WebhookAuth{
		logger:      logger,
		hookService: hookService,
	}
}

// IsAuthenticated checks if the request has a valid webhook token
func (m *WebhookAuth) IsAuthenticated(r *http.Request) bool {
	// Extract the hook ID from the URL path
	id := m.GetHookID(r)
	if id == "" {
		return false
	}

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		return false
	}

	// Validate hook token
	if err := m.hookService.ValidateHookToken(id, token); err != nil {
		return false
	}

	return true
}

// Middleware returns an http.Handler middleware function for webhook authentication
func (m *WebhookAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the hook ID from the URL path
		id := m.GetHookID(r)
		if id == "" {
			m.logger.Warn("Invalid webhook URL format",
				logger.Field{Key: "path", Value: r.URL.Path})
			http.Error(w, "Invalid webhook URL", http.StatusBadRequest)
			return
		}

		// Get token from query parameter
		token := r.URL.Query().Get("token")
		if token == "" {
			m.logger.Warn("Missing token parameter",
				logger.Field{Key: "id", Value: id})
			http.Error(w, "Missing token parameter", http.StatusBadRequest)
			return
		}

		// Validate hook token
		if err := m.hookService.ValidateHookToken(id, token); err != nil {
			if err == domain.ErrHookNotFound {
				m.logger.Warn("Hook not found",
					logger.Field{Key: "id", Value: id})
				http.Error(w, "Hook not found", http.StatusNotFound)
				return
			}
			if err == domain.ErrInvalidToken {
				m.logger.Warn("Invalid token",
					logger.Field{Key: "id", Value: id})
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			m.logger.Error("Failed to validate hook token",
				logger.Field{Key: "id", Value: id},
				logger.Field{Key: "error", Value: err.Error()})
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Call the next handler with webhook authenticated
		next.ServeHTTP(w, r)
	})
}

// GetHookID extracts hook ID from the URL path
// This is a helper function that can be used by handlers after webhook authentication
func (m *WebhookAuth) GetHookID(r *http.Request) string {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		return ""
	}
	return pathParts[len(pathParts)-1]
}
