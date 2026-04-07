package handler

import (
	"context"
	"encoding/json"
	"errors"
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
	tests := []struct {
		name                  string
		healthErr             error
		expectStatus          int
		expectHealthStatusVal string
		expectError           bool
	}{
		{
			name:                  "healthy",
			healthErr:             nil,
			expectStatus:          http.StatusOK,
			expectHealthStatusVal: "healthy",
			expectError:           false,
		},
		{
			name:                  "unhealthy",
			healthErr:             errors.New("storage unavailable"),
			expectStatus:          http.StatusServiceUnavailable,
			expectHealthStatusVal: "unhealthy",
			expectError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &MockHealthService{
				healthErr: tt.healthErr,
			}
			handler := NewHealthHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			handler.CheckHealth(w, req)

			if w.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, w.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp["status"] != tt.expectHealthStatusVal {
				t.Fatalf("expected health status value %q, got %q", tt.expectHealthStatusVal, resp["status"])
			}
			if _, ok := resp["uptime"]; !ok {
				t.Fatalf("expected uptime in response")
			}

			if tt.expectError {
				if errVal, ok := resp["error"]; !ok {
					t.Fatalf("expected error in response")
				} else if errVal != tt.healthErr.Error() {
					t.Fatalf("expected error message %q, got %q", tt.healthErr.Error(), errVal)
				}
			} else {
				if _, ok := resp["error"]; ok {
					t.Fatalf("unexpected error in health response")
				}
			}

		})
	}

}

func TestNewHealthHandler(t *testing.T) {
	svc := &MockHealthService{}
	handler := NewHealthHandler(svc)

	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	if handler.startTime.IsZero() {
		t.Fatal("expected startTime to be set")
	}
	// Verify startTime is recent (within 1 second)
	if time.Since(handler.startTime) > time.Second {
		t.Fatal("expected startTime to be recent")
	}
}
