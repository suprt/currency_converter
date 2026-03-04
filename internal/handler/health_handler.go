package handler

import (
	"context"
	"net/http"
	"time"
)

type HealthService interface {
	Health(ctx context.Context) error
}
type HealthHandler struct {
	service   HealthService
	startTime time.Time
}

func NewHealthHandler(service HealthService) *HealthHandler {
	return &HealthHandler{service: service, startTime: time.Now()}
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	err := h.service.Health(r.Context())
	status := "healthy"
	if err != nil {
		status = "unhealthy"
	}
	response := map[string]interface{}{
		"status":    status,
		"uptime":    time.Since(h.startTime).String(),
		"timestamp": time.Now().Unix(),
	}
	if err != nil {
		response["error"] = err.Error()
		JSON(w, http.StatusServiceUnavailable, response)
		return
	}
	JSON(w, http.StatusOK, response)
}
