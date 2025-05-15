package domain

import (
	"net/http"
)

// Middleware represents a generic HTTP middleware interface
type Middleware interface {
	// Middleware returns an http.Handler middleware function
	Middleware(next http.Handler) http.Handler
}

// AuthenticationMiddleware represents an authentication middleware
type AuthenticationMiddleware interface {
	Middleware
	// IsAuthenticated checks if the request is authenticated
	IsAuthenticated(r *http.Request) bool
}

// AdminAuthMiddleware provides authentication for admin API endpoints
type AdminAuthMiddleware interface {
	AuthenticationMiddleware
}

// WebhookAuthMiddleware provides authentication for webhook endpoints
type WebhookAuthMiddleware interface {
	AuthenticationMiddleware
	// GetHookID extracts hook ID from the request
	GetHookID(r *http.Request) string
}
