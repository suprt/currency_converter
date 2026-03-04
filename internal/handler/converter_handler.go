package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type ConverterService interface {
	GetRates(ctx context.Context, from, to string) (float64, error)
	Convert(ctx context.Context, from string, to string, amount float64) (float64, error)
}

type ConverterHandler struct {
	service ConverterService
}

func NewConverterHandler(service ConverterService) *ConverterHandler {
	return &ConverterHandler{service: service}
}

func (h *ConverterHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "Missing from and to parameters", http.StatusBadRequest)
		return
	}

	rate, err := h.service.GetRates(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"from": from,
		"to":   to,
		"rate": rate,
		"time": time.Now().Unix(),
	}
	JSON(w, http.StatusOK, response)
}

func (h *ConverterHandler) Convert(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	if from == "" || to == "" || amount <= 0 {
		http.Error(w, "Missing from, to or invalid amount", http.StatusBadRequest)
		return
	}

	result, err := h.service.Convert(r.Context(), from, to, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
		"result": result,
		"time":   time.Now().Unix(),
	}
	JSON(w, http.StatusOK, response)
}
