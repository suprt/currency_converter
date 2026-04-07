package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
		t.Fatalf("expected service to be set")
	}
}

func TestCurrencyHandler_List(t *testing.T) {
	tests := []struct {
		name              string
		data              []byte
		expectError       error
		expectStatus      int
		expectContentType string
	}{
		{
			name:              "success",
			data:              []byte(`{"EUR": "Euro", "GBP": "British Pound"}`),
			expectError:       nil,
			expectStatus:      http.StatusOK,
			expectContentType: "application/json",
		},
		{
			name:         "service error",
			expectError:  errors.New("failed to fetch currencies"),
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "empty response",
			data:         []byte(`{}`),
			expectError:  nil,
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &MockCurrenciesService{data: tt.data, err: tt.expectError}
			handler := NewCurrencyHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			if tt.expectError == nil {
				if w.Body.String() != string(tt.data) {
					t.Fatalf("expected  body %s, got %s", string(tt.data), w.Body.String())
				}
			} else {
				if !strings.Contains(w.Body.String(), tt.expectError.Error()) {
					t.Fatalf("expected  error message %q in body, got %q", tt.expectError.Error(), w.Body.String())
				}
			}

			if w.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, w.Code)
			}

			if tt.expectContentType != "" && w.Header().Get("Content-Type") != tt.expectContentType {
				t.Fatalf("expected Content-Type %s, got %s", tt.expectContentType, w.Header().Get("Content-Type"))
			}

		})
	}

}
