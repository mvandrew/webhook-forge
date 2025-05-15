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

// getHooks handles GET /api/hooks
func (h *Handler) getHooks(w http.ResponseWriter, r *http.Request) {
	hooks, err := h.hookService.GetAllHooks()
	if err != nil {
		h.logger.Error("Failed to get hooks", logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to get hooks")
		return
	}

	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hooks))
}

// getHook handles GET /api/hooks/{id}
func (h *Handler) getHook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	hook, err := h.hookService.GetHook(id)
	if err != nil {
		if err == domain.ErrHookNotFound {
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to get hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to get hook")
		return
	}

	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hook))
}

// createHook handles POST /api/hooks
func (h *Handler) createHook(w http.ResponseWriter, r *http.Request) {
	// Check admin token
	adminToken := r.Header.Get("X-Admin-Token")
	if adminToken == "" {
		h.logger.Warn("Missing admin token")
		h.respondError(w, http.StatusForbidden, "Admin token required")
		return
	}

	if adminToken != h.adminToken {
		h.logger.Warn("Invalid admin token")
		h.respondError(w, http.StatusForbidden, "Invalid admin token")
		return
	}

	var hook domain.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
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
		h.logger.Error("Failed to create hook", logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to create hook: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, domain.NewSuccessResponse(hook))
}

// updateHook handles PUT /api/hooks/{id}
func (h *Handler) updateHook(w http.ResponseWriter, r *http.Request) {
	// Check admin token
	adminToken := r.Header.Get("X-Admin-Token")
	if adminToken == "" {
		h.logger.Warn("Missing admin token")
		h.respondError(w, http.StatusForbidden, "Admin token required")
		return
	}

	if adminToken != h.adminToken {
		h.logger.Warn("Invalid admin token")
		h.respondError(w, http.StatusForbidden, "Invalid admin token")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	var hook domain.Hook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID matches path parameter
	hook.ID = id

	// Set update timestamp
	hook.UpdatedAt = time.Now()

	if err := h.hookService.UpdateHook(&hook); err != nil {
		if err == domain.ErrHookNotFound {
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to update hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to update hook: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(hook))
}

// deleteHook handles DELETE /api/hooks/{id}
func (h *Handler) deleteHook(w http.ResponseWriter, r *http.Request) {
	// Check admin token
	adminToken := r.Header.Get("X-Admin-Token")
	if adminToken == "" {
		h.logger.Warn("Missing admin token")
		h.respondError(w, http.StatusForbidden, "Admin token required")
		return
	}

	if adminToken != h.adminToken {
		h.logger.Warn("Invalid admin token")
		h.respondError(w, http.StatusForbidden, "Invalid admin token")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	if err := h.hookService.DeleteHook(id); err != nil {
		if err == domain.ErrHookNotFound {
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		h.logger.Error("Failed to delete hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to delete hook: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusNoContent, domain.NewSuccessResponse(nil))
}

// triggerHook handles POST /webhook/{id}
func (h *Handler) triggerHook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "Missing hook ID")
		return
	}

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		h.respondError(w, http.StatusBadRequest, "Missing token parameter")
		return
	}

	// Trigger hook
	if err := h.hookService.TriggerHook(id, token); err != nil {
		if err == domain.ErrHookNotFound {
			h.respondError(w, http.StatusNotFound, "Hook not found")
			return
		}
		if err == domain.ErrInvalidToken {
			h.respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		h.logger.Error("Failed to trigger hook", logger.Field{Key: "id", Value: id}, logger.Field{Key: "error", Value: err.Error()})
		h.respondError(w, http.StatusInternalServerError, "Failed to trigger hook: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(map[string]string{"status": "success"}))
}
