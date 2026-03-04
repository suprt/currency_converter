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

// List godoc
// @Summary List available currencies
// @Description Get a list of all available currencies with their codes and names
// @Tags currencies
// @Accept json
// @Produce json
// @Success 200 {string} json "List of currencies with codes and names"
// @Failure 500 {string} string "Internal server error"
// @Router /currencies [get]
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
