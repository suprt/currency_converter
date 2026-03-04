package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{})

		if limiter.rate != 10 {
			t.Errorf("expected default rate 10, got %d", limiter.rate)
		}
		if limiter.burst != 10 {
			t.Errorf("expected default burst 10, got %d", limiter.burst)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   20,
			Burst: 40,
		})

		if limiter.rate != 20 {
			t.Errorf("expected rate 20, got %d", limiter.rate)
		}
		if limiter.burst != 40 {
			t.Errorf("expected burst 40, got %d", limiter.burst)
		}
	})

	t.Run("negative RPS", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS: -5,
		})

		if limiter.rate != 10 {
			t.Errorf("expected default rate 10 for negative RPS, got %d", limiter.rate)
		}
	})

	t.Run("negative Burst", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   20,
			Burst: -5,
		})

		if limiter.burst != 20 {
			t.Errorf("expected burst equal to RPS for negative Burst, got %d", limiter.burst)
		}
	})
}

func TestTokenBucket_Limit(t *testing.T) {
	t.Run("allows requests within burst", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   10,
			Burst: 5,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := limiter.Limit(handler)

		// Should allow first 5 requests (burst)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			wrapped.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
			}
		}
	})

	t.Run("rejects requests exceeding burst", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   10,
			Burst: 3,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := limiter.Limit(handler)

		// Exhaust burst
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)
		}

		// Next request should be rate limited
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("expected status 429, got %d", w.Code)
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   10,
			Burst: 2,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := limiter.Limit(handler)

		// Exhaust burst for IP1
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "127.0.0.1:11111"
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)
		}

		// IP2 should still be allowed
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:22222"
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for different IP, got %d", w.Code)
		}
	})

	t.Run("uses X-Forwarded-For header", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   10,
			Burst: 2,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := limiter.Limit(handler)

		// Request with X-Forwarded-For
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestTokenBucket_Cleanup(t *testing.T) {
	limiter := NewTokenBucket(LimiterConfig{
		RPS:   10,
		Burst: 5,
	})

	// Add some buckets by making requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234" + string(rune(i))
		w := httptest.NewRecorder()
		handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		handler.ServeHTTP(w, req)
	}

	// Wait for buckets to become stale (more than 10 minutes)
	// Note: We can't actually wait 10 minutes in a test, so we'll just verify
	// that cleanup runs without errors
	limiter.Cleanup()

	// Buckets should still exist (not stale yet)
	limiter.mu.RLock()
	count := len(limiter.tokens)
	limiter.mu.RUnlock()

	if count != 3 {
		t.Errorf("expected 3 buckets, got %d", count)
	}
}

func TestTokenBucket_Start_Stop(t *testing.T) {
	limiter := NewTokenBucket(LimiterConfig{
		RPS:   10,
		Burst: 5,
	})

	// Start cleanup goroutine
	limiter.Start(100 * time.Millisecond)

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop should not panic
	limiter.Stop()

	// Give it time to stop
	time.Sleep(50 * time.Millisecond)
}

func TestTokenBucket_TokenRegeneration(t *testing.T) {
	limiter := NewTokenBucket(LimiterConfig{
		RPS:   10, // 10 tokens per second
		Burst: 2,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := limiter.Limit(handler)

	// Exhaust burst
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}

	// Should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Wait for token regeneration (100ms = 1 token at 10 RPS)
	time.Sleep(150 * time.Millisecond)

	// Should be allowed now
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 after token regeneration, got %d", w.Code)
	}
}
