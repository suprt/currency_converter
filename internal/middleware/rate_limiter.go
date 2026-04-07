package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type LimiterConfig struct {
	RPS       int
	Burst     int
	BucketTTL time.Duration
	KeyFunc   func(r *http.Request) string
}

type TokenBucket struct {
	rate      int
	burst     int
	bucketTTL time.Duration
	keyFunc   func(r *http.Request) string

	mu     sync.RWMutex
	tokens map[string]*bucket

	stopClean chan struct{}
}

type bucket struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewTokenBucket(cfg LimiterConfig) *TokenBucket {
	if cfg.RPS <= 0 {
		cfg.RPS = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.RPS
	}
	if cfg.BucketTTL <= 0 {
		cfg.BucketTTL = time.Minute * 10
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(r *http.Request) string {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.Header.Get("X-Real-Ip")
			}
			if ip == "" {
				ip = r.RemoteAddr
				if idx := strings.LastIndexByte(ip, ','); idx != -1 {
					ip = ip[:idx]
				}
			}
			return ip
		}
	}
	return &TokenBucket{
		rate:      cfg.RPS,
		burst:     cfg.Burst,
		bucketTTL: cfg.BucketTTL,
		keyFunc:   cfg.KeyFunc,
		tokens:    make(map[string]*bucket),
		stopClean: make(chan struct{}),
	}
}

func (t *TokenBucket) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := t.keyFunc(r)
		if !t.allow(key) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		}
		next.ServeHTTP(w, r)
	})
}

func (t *TokenBucket) allow(key string) bool {
	t.mu.Lock()
	b, ok := t.tokens[key]
	if !ok {
		b = &bucket{tokens: float64(t.burst), lastUpdate: time.Now()}
		t.tokens[key] = b

	}
	t.mu.Unlock()
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * float64(t.rate)
	if b.tokens > float64(t.burst) {
		b.tokens = float64(t.burst)
	}

	if b.tokens < 1 {
		return false
	}

	b.tokens -= 1
	b.lastUpdate = now
	return true
}

func (t *TokenBucket) Cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for key, b := range t.tokens {
		b.mu.Lock()

		if time.Since(b.lastUpdate) > t.bucketTTL {
			delete(t.tokens, key)
		}
		b.mu.Unlock()
	}
}

func (t *TokenBucket) Start(interval time.Duration) {
	// Если интервал не задан, используем 1 минуту по умолчанию
	if interval <= 0 {
		interval = 1 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.Cleanup()
			case <-t.stopClean:
				return
			}
		}
	}()
}

func (t *TokenBucket) Stop() {
	close(t.stopClean)

}
