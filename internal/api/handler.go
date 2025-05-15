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

// RegisterRoutes registers the API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// API routes with base path prefix
	apiPath := h.basePath + "/api"
	webhookPath := h.basePath + "/webhook"

	// Ensure paths are properly formatted
	apiPath = strings.TrimSuffix(apiPath, "/")
	webhookPath = strings.TrimSuffix(webhookPath, "/")

	// API routes
	mux.HandleFunc("GET "+apiPath+"/hooks", h.getHooks)
	mux.HandleFunc("GET "+apiPath+"/hooks/{id}", h.getHook)
	mux.HandleFunc("POST "+apiPath+"/hooks", h.createHook)
	mux.HandleFunc("PUT "+apiPath+"/hooks/{id}", h.updateHook)
	mux.HandleFunc("DELETE "+apiPath+"/hooks/{id}", h.deleteHook)

	// Webhook route
	mux.HandleFunc("POST "+webhookPath+"/{id}", h.triggerHook)

	h.logger.Info("Registered routes with base path", logger.Field{Key: "base_path", Value: h.basePath})
}

// getClientIP extracts the client IP address from the request, taking into account various headers
// that might be set by proxies or load balancers
func (h *Handler) getClientIP(r *http.Request) string {
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

// verifyAdminToken checks if the request has a valid admin token
func (h *Handler) verifyAdminToken(r *http.Request) bool {
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

	// Check admin token
	if !h.verifyAdminToken(r) {
		h.logger.Warn("Invalid or missing admin token",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusForbidden, "Admin authentication required")
		return
	}

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

	// Check admin token
	if !h.verifyAdminToken(r) {
		h.logger.Warn("Invalid or missing admin token",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusForbidden, "Admin authentication required")
		return
	}

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

	// Check admin token
	if !h.verifyAdminToken(r) {
		h.logger.Warn("Invalid or missing admin token",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusForbidden, "Admin authentication required")
		return
	}

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

	// Check admin token
	if !h.verifyAdminToken(r) {
		h.logger.Warn("Invalid or missing admin token",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusForbidden, "Admin authentication required")
		return
	}

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

	// Check admin token
	if !h.verifyAdminToken(r) {
		h.logger.Warn("Invalid or missing admin token",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusForbidden, "Admin authentication required")
		return
	}

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

	id := r.PathValue("id")
	if id == "" {
		h.logger.Warn("Missing hook ID in webhook request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "path", Value: r.URL.Path})
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		h.logger.Warn("Missing token parameter in webhook request",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "id", Value: id})
		h.respondError(w, http.StatusBadRequest, "Missing token parameter")
		return
	}

	// Trigger hook
	if err := h.hookService.TriggerHook(id, token, clientIP); err != nil {
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
