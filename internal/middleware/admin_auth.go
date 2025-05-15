package middleware

import (
	"net/http"
	"strings"

	"webhook-forge/internal/domain"
	"webhook-forge/pkg/logger"
)

// AdminAuth provides middleware for admin API endpoints authentication
type AdminAuth struct {
	logger     logger.Logger
	adminToken string
}

// NewAdminAuth creates a new admin authentication middleware
func NewAdminAuth(logger logger.Logger, adminToken string) domain.AdminAuthMiddleware {
	return &AdminAuth{
		logger:     logger,
		adminToken: adminToken,
	}
}

// IsAuthenticated checks if the request has a valid admin token
func (m *AdminAuth) IsAuthenticated(r *http.Request) bool {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")

	// Check if the header exists and has the correct format
	if authHeader == "" {
		return false
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return false
	}

	// Check if the token is valid
	token := parts[1]
	return token == m.adminToken
}

// Middleware returns an http.Handler middleware function for admin authentication
func (m *AdminAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is authenticated
		if !m.IsAuthenticated(r) {
			m.logger.Warn("Authentication failed",
				logger.Field{Key: "path", Value: r.URL.Path})
			http.Error(w, "Admin authentication required", http.StatusForbidden)
			return
		}

		// Call the next handler with admin authenticated
		next.ServeHTTP(w, r)
	})
}
