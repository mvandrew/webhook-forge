package api

import (
	"net/http"
	"time"

	"webhook-forge/internal/domain"
	"webhook-forge/pkg/logger"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// healthCheck handles GET /api/health
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Check hook service availability
	_, err := h.hookService.GetAllHooks()

	status := "up"
	if err != nil {
		status = "down"
		h.logger.Error("Health check failed",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "error", Value: err.Error()})
	} else {
		h.logger.Info("Health check succeeded",
			logger.Field{Key: "ip", Value: clientIP})
	}

	response := HealthStatus{
		Status:    status,
		Version:   "1.0.0", // Можно заменить на переменную с версией приложения
		Timestamp: time.Now(),
	}

	h.respondJSON(w, http.StatusOK, domain.NewSuccessResponse(response))
}
