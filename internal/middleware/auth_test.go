package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	t.Run("valid API key", func(t *testing.T) {
		auth := APIKeyAuth("secret-key-123")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := auth(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-KEY", "secret-key-123")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		auth := APIKeyAuth("secret-key-123")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := auth(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-KEY", "wrong-key")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing API key", func(t *testing.T) {
		auth := APIKeyAuth("secret-key-123")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := auth(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("empty API key in header", func(t *testing.T) {
		auth := APIKeyAuth("secret-key-123")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := auth(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-KEY", "")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("empty valid key", func(t *testing.T) {
		auth := APIKeyAuth("")

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := auth(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-API-KEY", "")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for empty valid key, got %d", w.Code)
		}
	})
}
