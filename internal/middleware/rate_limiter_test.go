package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tests := []struct {
		name              string
		rps               int
		burst             int
		bucketTTL         time.Duration
		expectedRPS       int
		expectedBurst     int
		expectedBucketTTL time.Duration
	}{
		{
			name:              "default values",
			expectedRPS:       10,
			expectedBurst:     10,
			expectedBucketTTL: 10 * time.Minute,
		},
		{
			name:              "custom values",
			rps:               20,
			burst:             40,
			bucketTTL:         1 * time.Minute,
			expectedRPS:       20,
			expectedBurst:     40,
			expectedBucketTTL: 1 * time.Minute,
		},
		{
			name:              "negative RPS",
			rps:               -1,
			expectedRPS:       10,
			expectedBurst:     10,
			expectedBucketTTL: 10 * time.Minute,
		},
		{
			name:              "negative Burst",
			rps:               20,
			burst:             -1,
			expectedRPS:       20,
			expectedBurst:     20,
			expectedBucketTTL: 10 * time.Minute,
		},
		{
			name:              "negative BucketTTL",
			expectedRPS:       10,
			expectedBurst:     10,
			bucketTTL:         -1 * time.Millisecond,
			expectedBucketTTL: 10 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewTokenBucket(LimiterConfig{RPS: tt.rps, Burst: tt.burst, BucketTTL: tt.bucketTTL})

			if limiter.rate != tt.expectedRPS {
				t.Fatalf("rps expected %v, got %v", tt.expectedRPS, limiter.rate)
			}
			if limiter.burst != tt.expectedBurst {
				t.Fatalf("burst expected %v, got %v", tt.expectedBurst, limiter.burst)
			}
			if limiter.bucketTTL != tt.expectedBucketTTL {
				t.Fatalf("bucketTTL expected %v, got %v", tt.expectedBucketTTL, limiter.bucketTTL)
			}
		})
	}
}

func TestTokenBucket_Limit(t *testing.T) {
	tests := []struct {
		name         string
		rps          int
		burst        int
		requests     int
		expectedCode int
	}{
		{
			name:         "allows requests within burst",
			rps:          10,
			burst:        5,
			requests:     5,
			expectedCode: http.StatusOK,
		},
		{
			name:         "reject requests exceeding burst",
			rps:          10,
			burst:        3,
			requests:     4,
			expectedCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewTokenBucket(LimiterConfig{
				RPS:   tt.rps,
				Burst: tt.burst,
			})

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := limiter.Limit(next)
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				wantStatus := http.StatusOK
				if i >= tt.burst {
					wantStatus = tt.expectedCode
				}
				if w.Code != wantStatus {
					t.Fatalf("req %d: expected %d, got %d", i+1, wantStatus, w.Code)
				}
			}

		})
	}

}
func TestTokenBucket_Limit_DifferentIPs(t *testing.T) {
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
			t.Fatalf("expected status 200 for different IP, got %d", w.Code)
		}
	})
}

func TestTokenBucket_Limit_Header(t *testing.T) {
	t.Run("uses X-Forwarded-For header", func(t *testing.T) {
		limiter := NewTokenBucket(LimiterConfig{
			RPS:   10,
			Burst: 2,
		})

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := limiter.Limit(next)

		// Request with X-Forwarded-For
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}

func TestTokenBucket_Cleanup(t *testing.T) {
	limiter := NewTokenBucket(LimiterConfig{
		RPS:       10,
		Burst:     5,
		BucketTTL: 100 * time.Millisecond,
	})
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	// Add some buckets by making requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234" + strconv.Itoa(i)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
	}

	if len(limiter.tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(limiter.tokens))
	}

	time.Sleep(200 * time.Millisecond)

	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.tokens)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Fatalf("expected buckets to be cleaned up")
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
		RPS:   10,
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
		t.Fatalf("expected status 429, got %d", w.Code)
	}

	// Wait for token regeneration (100ms = 1 token at 10 RPS)
	time.Sleep(150 * time.Millisecond)

	// Should be allowed now
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 after token regeneration, got %d", w.Code)
	}
}
