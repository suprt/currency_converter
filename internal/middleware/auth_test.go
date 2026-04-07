package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name       string
		validKey   string
		headerKey  *string
		wantStatus int
	}{
		{"valid API key", "secret-key-123", strPtr("secret-key-123"), http.StatusOK},
		{"invalid API key", "secret-key-123", strPtr("invalid-key"), http.StatusUnauthorized},
		{"missing header", "secret-key-123", nil, http.StatusUnauthorized},
		{"empty API key", "secret-key-123", strPtr(""), http.StatusUnauthorized},
		{"valid empty API key", "", strPtr(""), http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := APIKeyAuth(tt.validKey)
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := auth(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerKey != nil {
				req.Header.Set("X-API-KEY", *tt.headerKey)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if tt.wantStatus != w.Code {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}

}
