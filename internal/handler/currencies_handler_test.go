package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockCurrenciesService - mock for CurrenciesService interface
type MockCurrenciesService struct {
	data []byte
	err  error
}

func (m *MockCurrenciesService) GetCurrenciesJSON(ctx context.Context) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func TestNewCurrencyHandler(t *testing.T) {
	svc := &MockCurrenciesService{}
	handler := NewCurrencyHandler(svc)

	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	if handler.service != svc {
		t.Error("expected service to be set")
	}
}

func TestCurrencyHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &MockCurrenciesService{
			data: []byte(`{"EUR": "Euro", "GBP": "British Pound"}`),
		}
		handler := NewCurrencyHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		if string(w.Body.Bytes()) != `{"EUR": "Euro", "GBP": "British Pound"}` {
			t.Errorf("unexpected response body: %s", w.Body.String())
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &MockCurrenciesService{
			err: assertionError("failed to fetch currencies"),
		}
		handler := NewCurrencyHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		svc := &MockCurrenciesService{
			data: []byte(`{}`),
		}
		handler := NewCurrencyHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
