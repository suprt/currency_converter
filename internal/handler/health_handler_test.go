package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockHealthService - mock for HealthService interface
type MockHealthService struct {
	healthErr error
}

func (m *MockHealthService) Health(ctx context.Context) error {
	return m.healthErr
}

func TestHealthHandler_CheckHealth(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		svc := &MockHealthService{}
		handler := NewHealthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.CheckHealth(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["status"] != "healthy" {
			t.Errorf("expected status 'healthy', got '%v'", resp["status"])
		}
		if _, ok := resp["uptime"]; !ok {
			t.Error("expected uptime in response")
		}
		if _, ok := resp["timestamp"]; !ok {
			t.Error("expected timestamp in response")
		}
		if _, ok := resp["error"]; ok {
			t.Error("unexpected error in healthy response")
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		svc := &MockHealthService{
			healthErr: assertionError("storage unavailable"),
		}
		handler := NewHealthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.CheckHealth(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp["status"] != "unhealthy" {
			t.Errorf("expected status 'unhealthy', got '%v'", resp["status"])
		}
		if err, ok := resp["error"].(string); !ok || err != "storage unavailable" {
			t.Errorf("expected error message, got %v", resp["error"])
		}
	})
}

func TestNewHealthHandler(t *testing.T) {
	svc := &MockHealthService{}
	handler := NewHealthHandler(svc)

	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	if handler.startTime.IsZero() {
		t.Error("expected startTime to be set")
	}
	// Verify startTime is recent (within 1 second)
	if time.Since(handler.startTime) > time.Second {
		t.Error("expected startTime to be recent")
	}
}
