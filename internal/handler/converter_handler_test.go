package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// MockConverterService - mock for ConverterService interface
type MockConverterService struct {
	rates       map[string]float64
	convertErr  error
	getRatesErr error
}

func NewMockConverterService() *MockConverterService {
	return &MockConverterService{
		rates: make(map[string]float64),
	}
}

func (m *MockConverterService) GetRates(ctx context.Context, from, to string) (float64, error) {
	if m.getRatesErr != nil {
		return 0, m.getRatesErr
	}
	key := from + ":" + to
	if rate, ok := m.rates[key]; ok {
		return rate, nil
	}
	return 0.85, nil // Default rate
}

func (m *MockConverterService) Convert(ctx context.Context, from, to string, amount float64) (float64, error) {
	if m.convertErr != nil {
		return 0, m.convertErr
	}
	return amount * 0.85, nil // Default conversion
}

func (m *MockConverterService) SetRate(from, to string, rate float64) {
	m.rates[from+":"+to] = rate
}

func (m *MockConverterService) SetGetRatesError(err error) {
	m.getRatesErr = err
}

func (m *MockConverterService) SetConvertError(err error) {
	m.convertErr = err
}

func TestConverterHandler_GetRates(t *testing.T) {
	tests := []struct {
		name         string
		from         string
		to           string
		setupRate    bool
		rateVal      float64
		svcErr       error
		expectStatus int
		expectRate   float64
	}{
		{"success", "EUR", "GBP", true, 0.86, nil, http.StatusOK, 0.86},
		{"missing parameters", "", "", false, 0, nil, http.StatusBadRequest, 0},
		{"service error", "EUR", "GBP", true, 0.86, errors.New("service error"),
			http.StatusInternalServerError, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockConverterService()
			if tt.setupRate {
				svc.SetRate(tt.from, tt.to, tt.rateVal)
			}
			svc.getRatesErr = tt.svcErr
			handler := NewConverterHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rates?from=%s&to=%s", tt.from, tt.to), nil)
			w := httptest.NewRecorder()

			handler.GetRates(w, req)

			if tt.expectStatus != w.Code {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, w.Code)
			}

			if tt.expectStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if resp["from"] != tt.from || resp["to"] != tt.to {
					t.Fatalf("unexpected from/to in response")
				}
				if rate, ok := resp["rate"].(float64); !ok || rate != tt.expectRate {
					t.Fatalf("expected rate %f, got %v", tt.expectRate, resp["rate"])
				}
			}

		})
	}
}

func TestConverterHandler_Convert(t *testing.T) {
	tests := []struct {
		name         string
		from         string
		to           string
		amount       string
		svcErr       error
		expectStatus int
		expectResult float64
	}{
		{"success", "EUR", "GBP", "100", nil, http.StatusOK, 85},
		{"missing parameters", "", "", "", nil, http.StatusBadRequest, 0},
		{"invalid amount", "EUR", "GBP", "invalid", nil, http.StatusBadRequest, 0},
		{"negative amount", "EUR", "GBP", "-100", nil, http.StatusBadRequest, 0},
		{"zero amount", "EUR", "GBP", "0", nil, http.StatusBadRequest, 0},
		{"service error", "EUR", "GBP", "100", errors.New("conversion error"),
			http.StatusInternalServerError, 85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockConverterService()
			svc.convertErr = tt.svcErr
			handler := NewConverterHandler(svc)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/convert?from=%s&to=%s&amount=%s", tt.from, tt.to, tt.amount), nil)
			w := httptest.NewRecorder()

			handler.Convert(w, req)

			if tt.expectStatus != w.Code {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, w.Code)
			}
			if tt.expectStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if resp["from"] != tt.from || resp["to"] != tt.to {
					t.Fatalf("unexpected from/to in response")
				}
				s, err := strconv.ParseFloat(tt.amount, 64)
				if err != nil {
					t.Fatalf("failed to parse amount: %v", err)
				}
				if resp["amount"].(float64) != s {
					t.Fatalf("unexpected amount %f, got %f", s, resp["amount"].(float64))
				}
				if result, ok := resp["result"].(float64); !ok || result != tt.expectResult {
					t.Fatalf("expected result %f, got %f", result, resp["result"])
				}
			}
		})
	}
}

func TestConverterHandler_GetRates_InvalidCurrency(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		wantStatus int
	}{
		{"empty from", "", "USD", http.StatusBadRequest},
		{"short from", "EU", "USD", http.StatusBadRequest},
		{"long from", "EURO", "USD", http.StatusBadRequest},
		{"digits in from", "EU1", "USD", http.StatusBadRequest},
		{"lowercase", "eur", "usd", http.StatusOK},
		{"spec symbol in from", "EU^", "USD", http.StatusBadRequest},
		{"whitespace in around", "%20EUR", "USD", http.StatusOK},
		{"whitespace in param", "EU%20R", "USD", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockConverterService()
			handler := NewConverterHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/rates?from="+tt.from+"&to="+tt.to, nil)
			w := httptest.NewRecorder()
			handler.GetRates(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}

}

func TestConverterHandler_Convert_InvalidCurrency(t *testing.T) {
	tests := []struct {
		name       string
		from       string
		to         string
		amount     string
		wantStatus int
	}{
		{"empty from", "", "USD", "3", http.StatusBadRequest},
		{"short from", "EU", "USD", "3", http.StatusBadRequest},
		{"long from", "EURO", "USD", "3", http.StatusBadRequest},
		{"digits in from", "EU1", "USD", "3", http.StatusBadRequest},
		{"lowercase", "eur", "usd", "3", http.StatusOK},
		{"negative amount", "EUR", "USD", "-100", http.StatusBadRequest},
		{"zero amount", "EUR", "USD", "0", http.StatusBadRequest},
		{"letters in amount", "EUR", "USD", "abc", http.StatusBadRequest},
		{"spec symbol in from", "EU^", "USD", "3", http.StatusBadRequest},
		{"whitespace in around", "%20EUR", "USD", "3", http.StatusOK},
		{"whitespace in param", "EU%20R", "USD", "3", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockConverterService()
			svc.SetRate("EUR", "USD", 1.1)
			handler := NewConverterHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/convert?from="+tt.from+"&to="+tt.to+"&amount="+tt.amount, nil)
			w := httptest.NewRecorder()
			handler.Convert(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
