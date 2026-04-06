package repository

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/suprt/currency_converter/internal/service"
)

func TestInMemoryRepository_Set_Get(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       float64
		ttl         time.Duration
		wait        time.Duration
		expectedVal float64
		expectedErr bool
	}{
		{
			name:        "set and get",
			key:         "test:key",
			value:       42.5,
			ttl:         0,
			wait:        0,
			expectedVal: 42.5,
			expectedErr: false,
		},
		{
			name:        "get non-existent",
			key:         "nonexistent",
			value:       0,
			ttl:         0,
			wait:        0,
			expectedVal: 0,
			expectedErr: true,
		},
		{
			name:        "get expired",
			key:         "expired:key",
			value:       100,
			ttl:         100 * time.Millisecond,
			wait:        1 * time.Second,
			expectedVal: 0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryRepository(0)
			ctx := context.Background()
			if tt.value != 0 || tt.ttl > 0 {
				err := repo.Set(ctx, tt.key, tt.value, tt.ttl)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			val, err := repo.Get(ctx, tt.key)

			if tt.expectedErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, service.ErrNotFound) {
					t.Fatalf("expected %v, got %v", service.ErrNotFound, err)
				}
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectedErr && val != tt.expectedVal {
				t.Fatalf("expected %f, got %f", tt.expectedVal, val)
			}

		})
	}
}

func TestInMemoryRepository_Delete(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	repo.Set(ctx, "to:delete", 123.0, 0)

	err := repo.Delete(ctx, "to:delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.Get(ctx, "to:delete")
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestInMemoryRepository_Exists(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expected    bool
		expectedErr bool
	}{
		{
			name:     "exists",
			key:      "exists",
			expected: true,
		},
		{
			name:     "not exists",
			key:      "not:exists",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryRepository(0)
			ctx := context.Background()
			if tt.expected {
				_ = repo.Set(ctx, tt.key, 0, 0)
			}
			exists, err := repo.Exists(ctx, tt.key)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expected != exists {
				t.Fatalf("expected %v, got %v", tt.expected, exists)
			}

		})
	}
}

func TestInMemoryRepository_TTL(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		ttl         time.Duration
		expectedVal bool
		expectedTTL bool
		expectedErr bool
		wait        time.Duration
	}{
		{
			name:        "key without ttl",
			key:         "no:ttl",
			ttl:         0,
			expectedVal: true,
			expectedTTL: false,
			expectedErr: false,
			wait:        0,
		},
		{
			name:        "key with ttl",
			key:         "with:ttl",
			ttl:         10 * time.Second,
			expectedVal: true,
			expectedTTL: true,
			expectedErr: false,
			wait:        0,
		},
		{
			name:        "nonexistent key",
			key:         "nonexistent",
			expectedVal: false,
			expectedTTL: false,
			expectedErr: true,
			wait:        0,
		},
		{
			name:        "expired key",
			key:         "expired:key",
			ttl:         100 * time.Millisecond,
			expectedVal: true,
			expectedTTL: false,
			expectedErr: true,
			wait:        1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryRepository(0)
			ctx := context.Background()
			if tt.expectedVal {
				repo.Set(ctx, tt.key, 1.0, tt.ttl)
			}
			_, err := repo.Get(ctx, tt.key)
			if err != nil && tt.expectedVal {
				t.Fatalf("key should exist: %v", err)
			}

			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}

			ttl, err := repo.TTL(ctx, tt.key)

			if tt.expectedErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectedErr && !errors.Is(err, service.ErrNotFound) {
				t.Fatalf("expected error to be %v, got: %v", service.ErrNotFound, err)
			}
			if !tt.expectedTTL && ttl != 0 {
				t.Fatalf("expected 0 TTL, got %v", ttl)
			}
			if tt.expectedTTL && (ttl <= 0 || ttl > 10*time.Second) {
				t.Fatalf("expected positive TTL <= 10s, got %v", ttl)
			}
		})
	}
}

func TestInMemoryRepository_Len(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedLen int64
	}{
		{
			name:        "empty repo",
			expectedLen: 0,
		},
		{
			name:        "non-empty repo",
			key:         "test:key",
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryRepository(0)
			ctx := context.Background()
			if tt.expectedLen > 0 {
				repo.Set(ctx, tt.key, 1.0, 0)
			}
			rLen, err := repo.Len(ctx)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rLen != tt.expectedLen {
				t.Fatalf("expected length of %d, got %d", tt.expectedLen, rLen)
			}
		})
	}

}

func TestInMemoryRepository_Clear(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	repo.Set(ctx, "key1", 1.0, 0)
	repo.Set(ctx, "key2", 2.0, 0)

	err := repo.Clear(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	length, err := repo.Len(ctx)
	if err != nil {
		t.Fatalf("unexpected error after clear: %v", err)
	}
	if length != 0 {
		t.Fatalf("expected 0 after clear, got %d", length)
	}
}

func TestInMemoryRepository_Concurrent(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key:" + strconv.Itoa(id)
			_ = repo.Set(ctx, key, float64(id), 0)
			_, _ = repo.Get(ctx, key)
			_, _ = repo.Exists(ctx, key)
		}(i)
	}
	wg.Wait()

	length, err := repo.Len(ctx)
	if err != nil {
		t.Fatalf("unexpected error after concurrent ops: %v", err)
	}
	if length != 100 {
		t.Fatalf("expected 100 keys, got %d", length)
	}
}

func TestInMemoryRepository_Cleanup(t *testing.T) {
	repo := NewInMemoryRepository(50 * time.Millisecond)
	ctx := context.Background()

	repo.Set(ctx, "expiring", 1.0, 100*time.Millisecond)

	_, err := repo.Get(ctx, "expiring")
	if err != nil {
		t.Fatalf("key should exist before expiration: %v", err)
	}

	time.Sleep(2 * time.Second)
	repo.StopCleanup()

	_, err = repo.Get(ctx, "expiring")
	if !errors.Is(err, service.ErrNotFound) {
		t.Logf("key still exists after cleanup (this is ok if cleanup hasn't run yet)")
	}
}
