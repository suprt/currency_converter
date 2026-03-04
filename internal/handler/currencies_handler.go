package handler

import (
	"context"
	"net/http"
)

type CurrenciesService interface {
	GetCurrenciesJSON(ctx context.Context) ([]byte, error)
}
type CurrencyHandler struct {
	service CurrenciesService
}

func NewCurrencyHandler(service CurrenciesService) *CurrencyHandler {
	return &CurrencyHandler{
		service: service,
	}
}

func (h *CurrencyHandler) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetCurrenciesJSON(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
