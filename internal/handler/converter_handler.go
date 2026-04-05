package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
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

// GetRates godoc
// @Summary Get currency exchange rate
// @Description Get the exchange rate between two currencies
// @Tags conversion
// @Accept json
// @Produce json
// @Param from query string true "Source currency code (e.g., USD)"
// @Param to query string true "Target currency code (e.g., EUR)"
// @Success 200 {object} map[string]interface{} "Exchange rate response"
// @Failure 400 {string} string "Missing required parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /rates [get]
func (h *ConverterHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	from := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("from")))
	to := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("to")))
	if err := validateCurrencyCode(from); err != nil {
		http.Error(w, "Invalid 'from' parameter: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateCurrencyCode(to); err != nil {
		http.Error(w, "Invalid 'to' parameter: "+err.Error(), http.StatusBadRequest)
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

// Convert godoc
// @Summary Convert currency amount
// @Description Convert an amount from one currency to another
// @Tags conversion
// @Accept json
// @Produce json
// @Param from query string true "Source currency code (e.g., USD)"
// @Param to query string true "Target currency code (e.g., EUR)"
// @Param amount query number true "Amount to convert (must be positive)"
// @Success 200 {object} map[string]interface{} "Conversion result with original amount and converted result"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /convert [get]
func (h *ConverterHandler) Convert(w http.ResponseWriter, r *http.Request) {
	from := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("from")))
	to := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("to")))

	if err := validateCurrencyCode(from); err != nil {
		http.Error(w, "Invalid 'from' parameter: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateCurrencyCode(to); err != nil {
		http.Error(w, "Invalid 'to' parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	amountStr := strings.TrimSpace(r.URL.Query().Get("amount"))
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount: must be a number", http.StatusBadRequest)
		return
	}
	if amount <= 0 {
		http.Error(w, "Invalid amount: must be a positive number", http.StatusBadRequest)
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
