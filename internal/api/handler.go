package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"webhook-forge/internal/domain"
	"webhook-forge/pkg/logger"
)

// Handler handles HTTP requests
type Handler struct {
	hookService domain.HookService
	logger      logger.Logger
	basePath    string
	adminToken  string
}

// NewHandler creates a new handler
func NewHandler(hookService domain.HookService, logger logger.Logger, basePath string, adminToken string) *Handler {
	// Normalize base path: ensure it starts with '/' and doesn't end with '/'
	if basePath != "" {
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
	}

	return &Handler{
		hookService: hookService,
		logger:      logger,
		basePath:    basePath,
		adminToken:  adminToken,
	}
}

// GetAPIRoutes returns the API routes handler
func (h *Handler) GetAPIRoutes() http.Handler {
	apiMux := http.NewServeMux()

	// API routes
	apiMux.HandleFunc("GET /hooks", h.getHooks)
	apiMux.HandleFunc("GET /hooks/{id}", h.getHook)
	apiMux.HandleFunc("POST /hooks", h.createHook)
	apiMux.HandleFunc("PUT /hooks/{id}", h.updateHook)
	apiMux.HandleFunc("DELETE /hooks/{id}", h.deleteHook)

	return apiMux
}

// GetWebhookRoutes returns the webhook routes handler
func (h *Handler) GetWebhookRoutes() http.Handler {
	webhookMux := http.NewServeMux()

	// Webhook route
	webhookMux.HandleFunc("POST /{id}", h.triggerHook)

	return webhookMux
}

// getClientIP extracts the client IP address from the request
// Note: This is a temporary solution until we fully refactor the API handlers
// to use the middleware package for IP extraction
func (h *Handler) getClientIP(r *http.Request) string {
	// This implementation is now duplicated in middleware/logger.go
	// We'll keep it here temporarily for backward compatibility

	// Check X-Forwarded-For header first (common for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2, ...), take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (used by some proxies)
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr from the request
	// RemoteAddr is in the form "IP:port", so strip the port
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Remove brackets from IPv6 addresses
	ip = strings.TrimPrefix(ip, "[")
	ip = strings.TrimSuffix(ip, "]")

	return ip
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("Failed to encode response", logger.Field{Key: "error", Value: err.Error()})
		}
	}
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, domain.NewErrorResponse(message))
}

// verifyAdminToken is deprecated and will be removed
// Auth is now handled through middleware
func (h *Handler) verifyAdminToken(r *http.Request) bool {
	// This is kept temporarily for backward compatibility
	// Get Authorization header
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
	return token == h.adminToken
}

// getHooks handles GET /api/hooks
func (h *Handler) getHooks(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication is handled by middleware

	hooks, err := h.hookService.GetAllHooks()
	if err != nil {
		h.logger.Error("Failed to get hooks",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to get hooks")
		return
	}

	h.logger.Info("Hooks retrieved successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "count", Value: len(hooks)})
	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hooks))
}

// getHook handles GET /api/hooks/{id}
func (h *Handler) getHook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication is handled by middleware

	id := r.PathValue("id")
	if id == "" {
		h.logger.Warn("Missing hook ID in request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	hook, err := h.hookService.GetHook(id)
	if err != nil {
		if err == domain.ErrHookNotFound {
			h.logger.Warn("Hook not found",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "id", Value: id})
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to get hook",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to get hook")
		return
	}

	h.logger.Info("Hook retrieved successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "id", Value: id})
	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hook))
}

// createHook handles POST /api/hooks
func (h *Handler) createHook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication is handled by middleware

	var hook domain.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		h.logger.Warn("Invalid request body",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Generate token if not provided
	if hook.Token == "" {
		hook.Token = h.hookService.GenerateToken()
	}

	// Set timestamps
	now := time.Now()
	hook.CreatedAt = now
	hook.UpdatedAt = now

	if err := h.hookService.CreateHook(&hook); err != nil {
		h.logger.Error("Failed to create hook",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "name", Value: hook.Name},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to create hook: "+err.Error())
		return
	}

	h.logger.Info("Hook created successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "id", Value: hook.ID},
		logger.Field{Key: "name", Value: hook.Name})
	h.respondJSON(w, http.StatusCreated, domain.NewSuccessResponse(hook))
}

// updateHook handles PUT /api/hooks/{id}
func (h *Handler) updateHook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication is handled by middleware

	id := r.PathValue("id")
	if id == "" {
		h.logger.Warn("Missing hook ID in request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	var hook domain.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		h.logger.Warn("Invalid request body",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID matches path parameter
	hook.ID = id

	// Set update timestamp
	hook.UpdatedAt = time.Now()

	if err := h.hookService.UpdateHook(&hook); err != nil {
		if err == domain.ErrHookNotFound {
			h.logger.Warn("Hook not found",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "id", Value: id})
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to update hook",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to update hook: "+err.Error())
		return
	}

	h.logger.Info("Hook updated successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "id", Value: id})
	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hook))
}

// deleteHook handles DELETE /api/hooks/{id}
func (h *Handler) deleteHook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication is handled by middleware

	id := r.PathValue("id")
	if id == "" {
		h.logger.Warn("Missing hook ID in request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	if err := h.hookService.DeleteHook(id); err != nil {
		if err == domain.ErrHookNotFound {
			h.logger.Warn("Hook not found",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "id", Value: id})
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to delete hook",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to delete hook: "+err.Error())
		return
	}

	h.logger.Info("Hook deleted successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "id", Value: id})
	h.respondJSON(w, http.StatusNoContent, domain.NewSuccessResponse(nil))
}

// triggerHook handles POST /webhook/{id}
func (h *Handler) triggerHook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Authentication and ID extraction are handled by middleware
	id := r.PathValue("id")
	if id == "" {
		h.logger.Warn("Missing hook ID in webhook request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	// Get token from query parameter - already validated by middleware
	token := r.URL.Query().Get("token")

	// Trigger hook - token validation already done by middleware
	if err := h.hookService.TriggerHook(id, token, clientIP); err != nil {
		// These errors should not occur as they're handled by middleware
		// but we keep them for robustness
		if err == domain.ErrHookNotFound {
			h.logger.Warn("Hook not found in webhook request",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "id", Value: id})
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		if err == domain.ErrInvalidToken {
			h.logger.Warn("Invalid token in webhook request",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "id", Value: id})
			h.respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		h.logger.Error("Failed to trigger hook",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id},
			logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to trigger hook: "+err.Error())
		return
	}

	h.logger.Info("Hook triggered successfully",
		logger.Field{Key: "ip", Value: clientIP},
		logger.Field{Key: "id", Value: id})
	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(map[string]string{"status": "success"}))
}
