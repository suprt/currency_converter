package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	t.Run("success", func(t *testing.T) {
		svc := NewMockConverterService()
		svc.SetRate("EUR", "GBP", 0.86)
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/rates?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.GetRates(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["from"] != "EUR" || resp["to"] != "GBP" {
			t.Error("unexpected from/to in response")
		}
		if rate, ok := resp["rate"].(float64); !ok || rate != 0.86 {
			t.Errorf("expected rate 0.86, got %v", resp["rate"])
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/rates", nil)
		w := httptest.NewRecorder()

		handler.GetRates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing 'to' parameter", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/rates?from=EUR", nil)
		w := httptest.NewRecorder()

		handler.GetRates(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := NewMockConverterService()
		svc.SetGetRatesError(assertionError("service error"))
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/rates?from=EUR&to=GBP", nil)
		w := httptest.NewRecorder()

		handler.GetRates(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestConverterHandler_Convert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert?from=EUR&to=GBP&amount=100", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["from"] != "EUR" || resp["to"] != "GBP" {
			t.Error("unexpected from/to in response")
		}
		if resp["amount"].(float64) != 100 {
			t.Errorf("expected amount 100, got %v", resp["amount"])
		}
		if result, ok := resp["result"].(float64); !ok || result != 85 {
			t.Errorf("expected result 85, got %v", resp["result"])
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert?from=EUR&to=GBP&amount=invalid", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert?from=EUR&to=GBP&amount=-100", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		svc := NewMockConverterService()
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert?from=EUR&to=GBP&amount=0", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := NewMockConverterService()
		svc.SetConvertError(assertionError("conversion error"))
		handler := NewConverterHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/convert?from=EUR&to=GBP&amount=100", nil)
		w := httptest.NewRecorder()

		handler.Convert(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
