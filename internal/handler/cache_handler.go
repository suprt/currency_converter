package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type CacheService interface {
	DeleteRate(ctx context.Context, from, to string) error
	ExistsRate(ctx context.Context, from, to string) (bool, error)
	CacheSize(ctx context.Context) (int64, error)
	GetKey(ctx context.Context, from, to string) (float64, error)
	SetKey(ctx context.Context, from, to string, value float64, ttl time.Duration) error
	GetLastUpdate() time.Time
	Clear(ctx context.Context) error
	ForceRefresh(ctx context.Context) error
	TTL(ctx context.Context, from, to string) (time.Duration, error)
}

type CacheHandler struct {
	service CacheService
}

func NewCacheHandler(service CacheService) *CacheHandler {
	return &CacheHandler{
		service: service,
	}
}

func (h *CacheHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "Missing from and to parameters", http.StatusBadRequest)
		return
	}
	value, err := h.service.GetKey(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("%s:%s", from, to)
	JSON(w, http.StatusOK, map[string]float64{key: value})
}

func (h *CacheHandler) SetKey(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	value := r.URL.Query().Get("value")
	ttl := r.URL.Query().Get("ttl")
	if from == "" || to == "" || value == "" || ttl == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	ttlInt, err := strconv.Atoi(ttl)
	if err != nil || ttlInt <= 0 {
		http.Error(w, "Invalid ttl (must be positive integer in seconds)", http.StatusBadRequest)
		return
	}
	valueFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		http.Error(w, "Invalid value (must be a number)", http.StatusBadRequest)
		return
	}
	err = h.service.SetKey(r.Context(), from, to, valueFloat, time.Duration(ttlInt)*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CacheHandler) DeleteRate(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "Missing from and to parameters", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteRate(r.Context(), from, to); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CacheHandler) CacheSize(w http.ResponseWriter, r *http.Request) {
	size, err := h.service.CacheSize(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, http.StatusOK, map[string]interface{}{
		"size":       size,
		"lastUpdate": h.service.GetLastUpdate(),
	})
}

func (h *CacheHandler) CheckRate(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "Missing from and to parameters", http.StatusBadRequest)
		return
	}
	exists, err := h.service.ExistsRate(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, http.StatusOK, map[string]bool{"exists": exists})
}

func (h *CacheHandler) ClearAndRefresh(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Clear(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.service.ForceRefresh(r.Context()); err != nil {
		http.Error(w, "cache cleared but refresh failed: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CacheHandler) TTLKey(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "Missing from and to parameters", http.StatusBadRequest)
		return
	}
	ttl, err := h.service.TTL(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, http.StatusOK, map[string]interface{}{"from": from, "to": to, "ttl": ttl})
}
