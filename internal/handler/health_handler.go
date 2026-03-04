package handler

import (
	"context"
	"encoding/json"
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
	response := map[string]interface{}{"status": status, "uptime": time.Since(h.startTime).String(), "timestamp": time.Now().Unix()}
	if err != nil {
		response["error"] = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
