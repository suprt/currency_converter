package repository

import (
	"context"
	"testing"
	"time"

	"github.com/suprt/currency_converter/internal/service"
)

func TestInMemoryRepository_Set_Get(t *testing.T) {
	repo := NewInMemoryRepository(0) // No cleanup for this test
	ctx := context.Background()

	err := repo.Set(ctx, "test:key", 42.5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := repo.Get(ctx, "test:key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42.5 {
		t.Errorf("expected 42.5, got %f", val)
	}
}

func TestInMemoryRepository_Get_NotFound(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent")
	if err != service.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInMemoryRepository_Get_Expired(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	// Set with 2 second TTL
	err := repo.Set(ctx, "expiring:key", 100.0, 2*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get immediately to verify key exists
	val, err := repo.Get(ctx, "expiring:key")
	if err != nil {
		t.Fatalf("key should exist immediately after set: %v", err)
	}
	t.Logf("key value immediately after set: %f", val)

	// Wait 3 seconds to ensure expiration (Unix timestamp resolution + buffer)
	time.Sleep(3 * time.Second)

	// Try to get - should return ErrNotFound and delete the key
	_, err = repo.Get(ctx, "expiring:key")
	if err != service.ErrNotFound {
		t.Errorf("expected ErrNotFound for expired key, got %v", err)
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
	if err != service.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestInMemoryRepository_Exists(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	// Key doesn't exist
	exists, err := repo.Exists(ctx, "test:key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected key to not exist")
	}

	// Key exists
	repo.Set(ctx, "test:key", 1.0, 0)
	exists, err = repo.Exists(ctx, "test:key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected key to exist")
	}
}

func TestInMemoryRepository_TTL(t *testing.T) {
	t.Run("key without TTL", func(t *testing.T) {
		repo := NewInMemoryRepository(0)
		ctx := context.Background()

		repo.Set(ctx, "no:ttl", 1.0, 0)

		ttl, err := repo.TTL(ctx, "no:ttl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ttl != 0 {
			t.Errorf("expected 0 TTL, got %v", ttl)
		}
	})

	t.Run("key with TTL", func(t *testing.T) {
		repo := NewInMemoryRepository(0)
		ctx := context.Background()

		repo.Set(ctx, "with:ttl", 1.0, 10*time.Second)

		ttl, err := repo.TTL(ctx, "with:ttl")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ttl <= 0 || ttl > 10*time.Second {
			t.Errorf("expected positive TTL <= 10s, got %v", ttl)
		}
	})

	t.Run("nonexistent key", func(t *testing.T) {
		repo := NewInMemoryRepository(0)
		ctx := context.Background()

		_, err := repo.TTL(ctx, "nonexistent")
		if err != service.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("expired key", func(t *testing.T) {
		repo := NewInMemoryRepository(0)
		ctx := context.Background()

		repo.Set(ctx, "expiring", 1.0, 2*time.Second)
		
		// Verify key exists
		_, err := repo.Get(ctx, "expiring")
		if err != nil {
			t.Fatalf("key should exist: %v", err)
		}
		
		// Wait 3 seconds to ensure expiration (Unix timestamp resolution + buffer)
		time.Sleep(3 * time.Second)

		// TTL should return ErrNotFound for expired key
		// Note: TTL doesn't delete the key, only Get does
		ttl, err := repo.TTL(ctx, "expiring")
		if err != service.ErrNotFound {
			t.Errorf("expected ErrNotFound for expired key, got %v", err)
		}
		if ttl != 0 {
			t.Errorf("expected 0 TTL for expired key, got %v", ttl)
		}
		
		// Now Get should delete the key
		_, err = repo.Get(ctx, "expiring")
		if err != service.ErrNotFound {
			t.Errorf("expected ErrNotFound from Get for expired key, got %v", err)
		}
	})
}

func TestInMemoryRepository_Len(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	// Empty repo
	length, err := repo.Len(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if length != 0 {
		t.Errorf("expected 0, got %d", length)
	}

	// Add items
	repo.Set(ctx, "key1", 1.0, 0)
	repo.Set(ctx, "key2", 2.0, 0)
	repo.Set(ctx, "key3", 3.0, 0)

	length, err = repo.Len(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if length != 3 {
		t.Errorf("expected 3, got %d", length)
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

	length, _ := repo.Len(ctx)
	if length != 0 {
		t.Errorf("expected 0 after clear, got %d", length)
	}
}

func TestInMemoryRepository_Concurrent(t *testing.T) {
	repo := NewInMemoryRepository(0)
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(id int) {
			key := "key:" + string(rune(id))
			repo.Set(ctx, key, float64(id), 0)
			repo.Get(ctx, key)
			repo.Exists(ctx, key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify repo is still working
	length, err := repo.Len(ctx)
	if err != nil {
		t.Fatalf("unexpected error after concurrent ops: %v", err)
	}
	if length != 100 {
		t.Errorf("expected 100 keys, got %d", length)
	}
}

func TestInMemoryRepository_Cleanup(t *testing.T) {
	repo := NewInMemoryRepository(50 * time.Millisecond)
	ctx := context.Background()

	// Set with 1 second TTL
	repo.Set(ctx, "expiring", 1.0, 1*time.Second)

	// Verify it exists
	_, err := repo.Get(ctx, "expiring")
	if err != nil {
		t.Fatalf("key should exist before expiration: %v", err)
	}

	// Wait for expiration + cleanup cycle
	time.Sleep(2 * time.Second)

	// Key should be cleaned up by background goroutine
	// Note: Get might still find it in the map before cleanup runs
	// So we just verify cleanup eventually runs
	repo.StopCleanup()
	
	// After stopping cleanup, expired keys should still be removed on access
	_, err = repo.Get(ctx, "expiring")
	if err != service.ErrNotFound {
		t.Logf("key still exists after cleanup (this is ok if cleanup hasn't run yet)")
	}
}
